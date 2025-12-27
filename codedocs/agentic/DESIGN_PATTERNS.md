# Agentic Design Patterns Reference

**Source:** [Agentic Design Patterns by Antonio Gulli](https://github.com/sarwarbeing-ai/Agentic_Design_Patterns)
**Repository:** 6.1k stars, 1.2k forks

---

## Core Patterns (Chapters 1-5)

### 1. Prompt Chaining
Sequential composition of prompts where output of one becomes input to the next.
- **Use Case:** Complex reasoning tasks that benefit from step-by-step decomposition
- **Example:** Extract → Transform → Validate → Format

### 2. Routing
Dynamic selection of processing paths based on input characteristics.
- **Use Case:** Classifying requests and routing to specialized handlers
- **Implementations:** Google ADK, LangGraph, Openrouter

### 3. Parallelization
Concurrent execution of independent tasks for efficiency.
- **Use Case:** Processing multiple documents, API calls, or analyses simultaneously
- **Implementations:** Google ADK, LangChain

### 4. Reflection
Self-evaluation and iterative improvement of outputs.
- **Use Case:** Quality assurance, self-correction, iterative refinement
- **Variants:** Single-pass, Iterative Loop, Multi-agent critique

### 5. Tool Use
Integration with external tools and APIs for enhanced capabilities.
- **Use Case:** Code execution, web search, database queries, file operations
- **Implementations:** CrewAI, LangChain, Vertex AI Search, Google Search

---

## Advanced Patterns (Chapters 6-10)

### 6. Planning
Strategic decomposition of goals into actionable steps.
- **Use Case:** Complex task execution, research automation
- **Variants:** Code-based planning, Deep Research API

### 7. Multi-Agent Collaboration
Coordination between multiple specialized agents.
- **Use Case:** Complex problem-solving requiring diverse expertise
- **Implementations:** 5 ADK variants, CrewAI, Gemini-based

### 8. Memory Management
Persistent storage and retrieval of context across interactions.
- **Use Case:** Long conversations, user preferences, learned patterns
- **Types:** Short-term, Long-term, Episodic, Semantic, Procedural

### 9. Adaptation
Dynamic adjustment of behavior based on feedback and results.
- **Use Case:** Learning from mistakes, optimizing performance
- **Example:** OpenEvolve pattern

### 10. Model Context Protocol (MCP)
Standardized interface for model-tool communication.
- **Use Case:** Extensible tool integration, consistent API patterns
- **Components:** FastMCP Server, Client Agent, Filesystem operations

---

## Operational Patterns (Chapters 11-21)

### 11. Goal Setting and Monitoring
Defining objectives and tracking progress toward completion.
- **Use Case:** Task management, success metrics, milestone tracking

### 12. Exception Handling and Recovery
Graceful handling of errors and failure recovery strategies.
- **Use Case:** Robustness, retry logic, fallback behaviors

### 13. Human-in-the-Loop
Integration of human oversight and intervention points.
- **Use Case:** High-stakes decisions, approval workflows, quality control
- **Example:** Customer Support Agent with escalation

### 14. Knowledge Retrieval (RAG)
Retrieval-Augmented Generation for grounded responses.
- **Use Case:** Factual Q&A, document analysis, knowledge bases
- **Implementations:** LangChain RAG, Google Search, VertexAI

### 15. Inter-Agent Communication (A2A)
Protocols for agents to communicate and coordinate.
- **Use Case:** Multi-agent systems, task delegation, information sharing
- **Variants:** Sync, Streaming, Event-based (WeatherBot example)

### 16. Resource-Aware Optimization
Efficient use of compute, memory, and API resources.
- **Use Case:** Cost management, latency optimization, scaling

### 17. Reasoning Techniques
Structured approaches to problem-solving and inference.
- **Types:**
  - Chain-of-Thought (CoT)
  - Self-correction
  - Code Execution
  - DeepSearch

### 18. Guardrails & Safety Patterns
Constraints and validation for safe agent behavior.
- **Use Case:** Content filtering, action validation, boundary enforcement
- **Components:** Input validation, Output filtering, Action constraints, Audit logging

### 19. Evaluation and Monitoring
Assessment of agent performance and behavior.
- **Methods:**
  - Response Evaluation
  - LLM-as-Judge
  - Metrics collection
  - Observability

### 20. Prioritization
Ranking and scheduling of tasks by importance.
- **Use Case:** Resource allocation, workload management
- **Example:** SuperSimplePM (Priority Manager)

### 21. Exploration and Discovery
Autonomous investigation and learning behaviors.
- **Use Case:** Research automation, hypothesis generation
- **Example:** Agent Laboratory pattern

---

## Pattern Selection Guide

| Scenario | Recommended Patterns |
|----------|---------------------|
| Complex reasoning | Prompt Chaining + Reflection |
| Multi-step tasks | Planning + Tool Use |
| High reliability | Exception Handling + Guardrails |
| Knowledge work | RAG + Memory Management |
| Team of agents | Multi-Agent + A2A + Prioritization |
| Production systems | Monitoring + Resource-Aware + Safety |

---

## JSON Schema

```json
{
  "patterns": [
    {
      "id": "prompt-chaining",
      "category": "core",
      "chapter": 1,
      "name": "Prompt Chaining",
      "description": "Sequential composition of prompts",
      "useCases": ["complex reasoning", "step-by-step decomposition"]
    },
    {
      "id": "routing",
      "category": "core",
      "chapter": 2,
      "name": "Routing",
      "description": "Dynamic selection of processing paths",
      "useCases": ["request classification", "specialized handlers"]
    },
    {
      "id": "parallelization",
      "category": "core",
      "chapter": 3,
      "name": "Parallelization",
      "description": "Concurrent execution of independent tasks",
      "useCases": ["batch processing", "efficiency"]
    },
    {
      "id": "reflection",
      "category": "core",
      "chapter": 4,
      "name": "Reflection",
      "description": "Self-evaluation and iterative improvement",
      "useCases": ["quality assurance", "self-correction"]
    },
    {
      "id": "tool-use",
      "category": "core",
      "chapter": 5,
      "name": "Tool Use",
      "description": "Integration with external tools and APIs",
      "useCases": ["code execution", "web search", "file operations"]
    },
    {
      "id": "planning",
      "category": "advanced",
      "chapter": 6,
      "name": "Planning",
      "description": "Strategic decomposition into actionable steps",
      "useCases": ["complex tasks", "research automation"]
    },
    {
      "id": "multi-agent",
      "category": "advanced",
      "chapter": 7,
      "name": "Multi-Agent Collaboration",
      "description": "Coordination between specialized agents",
      "useCases": ["diverse expertise", "complex problem-solving"]
    },
    {
      "id": "memory",
      "category": "advanced",
      "chapter": 8,
      "name": "Memory Management",
      "description": "Persistent context storage and retrieval",
      "useCases": ["long conversations", "user preferences"]
    },
    {
      "id": "adaptation",
      "category": "advanced",
      "chapter": 9,
      "name": "Adaptation",
      "description": "Dynamic behavior adjustment from feedback",
      "useCases": ["learning", "optimization"]
    },
    {
      "id": "mcp",
      "category": "advanced",
      "chapter": 10,
      "name": "Model Context Protocol",
      "description": "Standardized model-tool communication",
      "useCases": ["tool integration", "extensibility"]
    },
    {
      "id": "goal-monitoring",
      "category": "operational",
      "chapter": 11,
      "name": "Goal Setting and Monitoring",
      "description": "Objective definition and progress tracking",
      "useCases": ["task management", "milestones"]
    },
    {
      "id": "exception-handling",
      "category": "operational",
      "chapter": 12,
      "name": "Exception Handling",
      "description": "Error handling and recovery strategies",
      "useCases": ["robustness", "retry logic"]
    },
    {
      "id": "human-in-loop",
      "category": "operational",
      "chapter": 13,
      "name": "Human-in-the-Loop",
      "description": "Human oversight and intervention",
      "useCases": ["high-stakes decisions", "approvals"]
    },
    {
      "id": "rag",
      "category": "operational",
      "chapter": 14,
      "name": "Knowledge Retrieval (RAG)",
      "description": "Retrieval-augmented generation",
      "useCases": ["factual Q&A", "document analysis"]
    },
    {
      "id": "a2a",
      "category": "operational",
      "chapter": 15,
      "name": "Inter-Agent Communication",
      "description": "Agent-to-agent protocols",
      "useCases": ["multi-agent systems", "task delegation"]
    },
    {
      "id": "resource-aware",
      "category": "operational",
      "chapter": 16,
      "name": "Resource-Aware Optimization",
      "description": "Efficient resource utilization",
      "useCases": ["cost management", "scaling"]
    },
    {
      "id": "reasoning",
      "category": "operational",
      "chapter": 17,
      "name": "Reasoning Techniques",
      "description": "Structured problem-solving approaches",
      "useCases": ["chain-of-thought", "self-correction"]
    },
    {
      "id": "guardrails",
      "category": "operational",
      "chapter": 18,
      "name": "Guardrails & Safety",
      "description": "Constraints for safe agent behavior",
      "useCases": ["content filtering", "validation"]
    },
    {
      "id": "evaluation",
      "category": "operational",
      "chapter": 19,
      "name": "Evaluation and Monitoring",
      "description": "Performance assessment",
      "useCases": ["metrics", "observability"]
    },
    {
      "id": "prioritization",
      "category": "operational",
      "chapter": 20,
      "name": "Prioritization",
      "description": "Task ranking and scheduling",
      "useCases": ["resource allocation", "workload management"]
    },
    {
      "id": "exploration",
      "category": "operational",
      "chapter": 21,
      "name": "Exploration and Discovery",
      "description": "Autonomous investigation",
      "useCases": ["research automation", "hypothesis generation"]
    }
  ]
}
```

---

**Last Updated:** 2025-12-27
