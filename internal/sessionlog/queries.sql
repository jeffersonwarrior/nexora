-- Nexora Session Logging: Useful Queries
-- Run these against the nexora_sessions database to analyze multi-instance behavior

-- =====================================================
-- REAL-TIME DASHBOARD QUERIES
-- =====================================================

-- 1. Current active sessions across all instances
SELECT 
  instance_id,
  COUNT(*) as active_sessions,
  MIN(started_at) as oldest_started,
  MAX(started_at) as newest_started
FROM nexora_sessions
WHERE status = 'active'
GROUP BY instance_id
ORDER BY active_sessions DESC;

-- 2. Edit success/failure rate by instance (last 1 hour)
SELECT 
  instance_id,
  COUNT(*) as total_edits,
  SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as successes,
  SUM(CASE WHEN status = 'failure' THEN 1 ELSE 0 END) as failures,
  ROUND(100.0 * SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) / COUNT(*), 2) as success_rate
FROM nexora_edit_operations
WHERE timestamp > NOW() - INTERVAL '1 hour'
GROUP BY instance_id
ORDER BY success_rate ASC;

-- 3. Most common edit failure reasons
SELECT 
  failure_reason,
  COUNT(*) as occurrences,
  ROUND(100.0 * COUNT(*) / (SELECT COUNT(*) FROM nexora_edit_operations WHERE status = 'failure' AND timestamp > NOW() - INTERVAL '1 hour'), 2) as pct
FROM nexora_edit_operations
WHERE status = 'failure' AND timestamp > NOW() - INTERVAL '1 hour'
GROUP BY failure_reason
ORDER BY occurrences DESC
LIMIT 10;

-- 4. Problematic files (most failures)
SELECT 
  file_path,
  instance_id,
  COUNT(*) as failures,
  ROUND(AVG(attempt_count), 2) as avg_attempts
FROM nexora_edit_operations
WHERE status = 'failure' AND timestamp > NOW() - INTERVAL '24 hours'
GROUP BY file_path, instance_id
ORDER BY failures DESC
LIMIT 20;

-- 5. Whitespace-related failures
SELECT 
  instance_id,
  COUNT(*) as whitespace_failures,
  SUM(CASE WHEN has_tabs THEN 1 ELSE 0 END) as tab_issues,
  SUM(CASE WHEN has_mixed_indent THEN 1 ELSE 0 END) as mixed_indent_issues,
  COUNT(DISTINCT CASE WHEN file_line_endings = 'CRLF' THEN file_path END) as crlf_files,
  COUNT(DISTINCT CASE WHEN file_line_endings = 'Mixed' THEN file_path END) as mixed_endings_files
FROM nexora_edit_operations
WHERE status = 'failure' AND (has_tabs OR has_mixed_indent OR file_line_endings IN ('CRLF', 'Mixed'))
AND timestamp > NOW() - INTERVAL '24 hours'
GROUP BY instance_id
ORDER BY whitespace_failures DESC;

-- =====================================================
-- PERFORMANCE ANALYSIS
-- =====================================================

-- 6. Edit performance by file size
SELECT 
  CASE 
    WHEN old_string_length < 100 THEN '<100'
    WHEN old_string_length < 500 THEN '100-500'
    WHEN old_string_length < 1000 THEN '500-1K'
    WHEN old_string_length < 5000 THEN '1K-5K'
    ELSE '>5K'
  END as edit_size_bucket,
  COUNT(*) as total_edits,
  ROUND(AVG(duration_ms), 2) as avg_duration_ms,
  MAX(duration_ms) as max_duration_ms,
  ROUND(100.0 * SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) / COUNT(*), 2) as success_rate
FROM nexora_edit_operations
WHERE timestamp > NOW() - INTERVAL '24 hours'
GROUP BY edit_size_bucket
ORDER BY old_string_length;

-- 7. View operation performance
SELECT 
  instance_id,
  COUNT(*) as view_count,
  ROUND(AVG(duration_ms), 2) as avg_duration_ms,
  ROUND(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms), 2) as p95_duration_ms,
  MAX(duration_ms) as max_duration_ms
FROM nexora_view_operations
WHERE timestamp > NOW() - INTERVAL '24 hours'
GROUP BY instance_id
ORDER BY avg_duration_ms DESC;

-- 8. Slowest operations across all instances
SELECT 
  instance_id,
  'edit' as operation_type,
  file_path,
  duration_ms,
  status,
  timestamp
FROM nexora_edit_operations
WHERE timestamp > NOW() - INTERVAL '24 hours'
UNION ALL
SELECT 
  instance_id,
  'view' as operation_type,
  file_path,
  duration_ms,
  status,
  timestamp
FROM nexora_view_operations
WHERE timestamp > NOW() - INTERVAL '24 hours'
ORDER BY duration_ms DESC
LIMIT 20;

-- =====================================================
-- TROUBLESHOOTING & ROOT CAUSE ANALYSIS
-- =====================================================

-- 9. Sessions with high error counts
SELECT 
  session_id,
  instance_id,
  started_at,
  ROUND(EXTRACT(EPOCH FROM (COALESCE(ended_at, NOW()) - started_at)), 2) as duration_seconds,
  error_count,
  tool_count,
  ROUND(100.0 * error_count / NULLIF(tool_count, 0), 2) as error_rate,
  status
FROM nexora_sessions
WHERE error_count > 5 AND created_at > NOW() - INTERVAL '7 days'
ORDER BY error_count DESC
LIMIT 20;

-- 10. Instance comparison: edits by status
SELECT 
  instance_id,
  status,
  COUNT(*) as count,
  ROUND(AVG(duration_ms), 2) as avg_duration_ms,
  ROUND(AVG(attempt_count), 2) as avg_attempts,
  COUNT(DISTINCT session_id) as unique_sessions
FROM nexora_edit_operations
WHERE timestamp > NOW() - INTERVAL '7 days'
GROUP BY instance_id, status
ORDER BY instance_id, status;

-- 11. File-specific failure analysis
SELECT 
  file_path,
  COUNT(*) as total_edits,
  SUM(CASE WHEN status = 'failure' THEN 1 ELSE 0 END) as failures,
  ROUND(100.0 * SUM(CASE WHEN status = 'failure' THEN 1 ELSE 0 END) / COUNT(*), 2) as failure_rate,
  ROUND(AVG(attempt_count), 2) as avg_attempts,
  STRING_AGG(DISTINCT failure_reason, ', ') as failure_types
FROM nexora_edit_operations
WHERE timestamp > NOW() - INTERVAL '7 days'
GROUP BY file_path
HAVING COUNT(*) >= 5
ORDER BY failure_rate DESC
LIMIT 30;

-- 12. Instance health snapshot
SELECT 
  instance_id,
  (SELECT COUNT(*) FROM nexora_sessions WHERE instance_id = i.instance_id AND status = 'active') as active_sessions,
  (SELECT COUNT(*) FROM nexora_edit_operations WHERE instance_id = i.instance_id AND timestamp > NOW() - INTERVAL '1 hour') as edits_last_hour,
  (SELECT ROUND(100.0 * SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) / NULLIF(COUNT(*), 0), 2) 
   FROM nexora_edit_operations WHERE instance_id = i.instance_id AND timestamp > NOW() - INTERVAL '1 hour') as success_rate_1h,
  (SELECT COUNT(*) FROM nexora_view_operations WHERE instance_id = i.instance_id AND timestamp > NOW() - INTERVAL '1 hour') as views_last_hour
FROM (
  SELECT DISTINCT instance_id FROM nexora_sessions
  UNION
  SELECT DISTINCT instance_id FROM nexora_edit_operations
  UNION
  SELECT DISTINCT instance_id FROM nexora_view_operations
) i
ORDER BY instance_id;
