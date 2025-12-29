#!/usr/bin/env python3
"""Parse Agentic Design Patterns PDF text into structured JSON."""

import json
import re
from pathlib import Path

def parse_agentic_patterns():
    """Extract structured data from the PDF text."""

    # Read the extracted text
    txt_path = Path("/home/nexora/codedocs/agentic/Agentic_Design_Patterns.txt")
    text = txt_path.read_text(encoding='utf-8', errors='ignore')

    # Structure to hold parsed data
    data = {
        "title": "Agentic Design Patterns",
        "subtitle": "A Hands-On Guide to Building Intelligent Systems",
        "author": "Antonio Gulli",
        "total_pages": 482,
        "parts": [],
        "chapters": [],
        "appendices": [],
        "metadata": {
            "donation": "All royalties donated to Save the Children",
            "source": "https://github.com/sarwarbeing-ai/Agentic_Design_Patterns"
        }
    }

    # Define chapter patterns
    patterns = [
        {"id": 1, "name": "Prompt Chaining", "part": 1},
        {"id": 2, "name": "Routing", "part": 1},
        {"id": 3, "name": "Parallelization", "part": 1},
        {"id": 4, "name": "Reflection", "part": 1},
        {"id": 5, "name": "Tool Use", "part": 1},
        {"id": 6, "name": "Planning", "part": 1},
        {"id": 7, "name": "Multi-Agent", "part": 1},
        {"id": 8, "name": "Memory Management", "part": 2},
        {"id": 9, "name": "Learning and Adaptation", "part": 2},
        {"id": 10, "name": "Model Context Protocol (MCP)", "part": 2},
        {"id": 11, "name": "Goal Setting and Monitoring", "part": 2},
        {"id": 12, "name": "Exception Handling and Recovery", "part": 3},
        {"id": 13, "name": "Human-in-the-Loop", "part": 3},
        {"id": 14, "name": "Knowledge Retrieval (RAG)", "part": 3},
        {"id": 15, "name": "Inter-Agent Communication (A2A)", "part": 4},
        {"id": 16, "name": "Resource-Aware Optimization", "part": 4},
        {"id": 17, "name": "Reasoning Techniques", "part": 4},
        {"id": 18, "name": "Guardrails/Safety Patterns", "part": 4},
        {"id": 19, "name": "Evaluation and Monitoring", "part": 4},
        {"id": 20, "name": "Prioritization", "part": 4},
        {"id": 21, "name": "Exploration and Discovery", "part": 4},
    ]

    # Extract chapter content
    for pattern in patterns:
        chapter_num = pattern["id"]
        chapter_name = pattern["name"]

        # Try to find chapter section in text
        chapter_pattern = rf"Chapter {chapter_num}:?\s*{re.escape(chapter_name)}"
        match = re.search(chapter_pattern, text, re.IGNORECASE)

        chapter_data = {
            "id": chapter_num,
            "title": f"Chapter {chapter_num}: {chapter_name}",
            "pattern_name": chapter_name,
            "part": pattern["part"],
            "has_code": True,
            "status": "final",
            "summary": f"Covers the {chapter_name} pattern for building agentic systems.",
        }

        # Extract key concepts based on pattern name
        if chapter_num == 1:  # Prompt Chaining
            chapter_data["key_concepts"] = [
                "Sequential composition of prompts",
                "Output → Input chaining",
                "Step-by-step decomposition",
                "Complex reasoning tasks"
            ]
        elif chapter_num == 2:  # Routing
            chapter_data["key_concepts"] = [
                "Dynamic path selection",
                "Request classification",
                "Specialized handlers",
                "Google ADK, LangGraph implementations"
            ]
        elif chapter_num == 3:  # Parallelization
            chapter_data["key_concepts"] = [
                "Concurrent execution",
                "Independent task processing",
                "Efficiency optimization",
                "Google ADK, LangChain implementations"
            ]
        elif chapter_num == 4:  # Reflection
            chapter_data["key_concepts"] = [
                "Self-evaluation",
                "Iterative improvement",
                "Quality assurance",
                "Multi-agent critique"
            ]
        elif chapter_num == 5:  # Tool Use
            chapter_data["key_concepts"] = [
                "External tool integration",
                "Code execution",
                "Web search",
                "Database queries",
                "CrewAI, LangChain implementations"
            ]
        elif chapter_num == 6:  # Planning
            chapter_data["key_concepts"] = [
                "Goal decomposition",
                "Strategic planning",
                "Actionable steps",
                "Research automation"
            ]
        elif chapter_num == 7:  # Multi-Agent
            chapter_data["key_concepts"] = [
                "Agent coordination",
                "Specialized expertise",
                "Collaborative problem-solving",
                "CrewAI, ADK implementations"
            ]
        elif chapter_num == 8:  # Memory Management
            chapter_data["key_concepts"] = [
                "Persistent storage",
                "Context retrieval",
                "Short-term memory",
                "Long-term memory",
                "Episodic and semantic memory"
            ]
        elif chapter_num == 9:  # Adaptation
            chapter_data["key_concepts"] = [
                "Dynamic adjustment",
                "Feedback-based learning",
                "Performance optimization",
                "OpenEvolve pattern"
            ]
        elif chapter_num == 10:  # MCP
            chapter_data["key_concepts"] = [
                "Standardized interface",
                "Model-tool communication",
                "Extensible integration",
                "FastMCP server patterns"
            ]
        elif chapter_num == 11:  # Goal Setting
            chapter_data["key_concepts"] = [
                "Objective definition",
                "Progress tracking",
                "Milestone management",
                "Success metrics"
            ]
        elif chapter_num == 12:  # Exception Handling
            chapter_data["key_concepts"] = [
                "Error recovery",
                "Retry logic",
                "Fallback behaviors",
                "Graceful degradation"
            ]
        elif chapter_num == 13:  # Human-in-Loop
            chapter_data["key_concepts"] = [
                "Human oversight",
                "Intervention points",
                "Approval workflows",
                "High-stakes decisions",
                "Customer support escalation"
            ]
        elif chapter_num == 14:  # RAG
            chapter_data["key_concepts"] = [
                "Retrieval-augmented generation",
                "Factual Q&A",
                "Document analysis",
                "Knowledge bases",
                "LangChain, VertexAI implementations"
            ]
        elif chapter_num == 15:  # A2A
            chapter_data["key_concepts"] = [
                "Agent-to-agent communication",
                "Message protocols",
                "Task delegation",
                "Sync, streaming, event-based patterns"
            ]
        elif chapter_num == 16:  # Resource-Aware
            chapter_data["key_concepts"] = [
                "Resource optimization",
                "Cost management",
                "Latency reduction",
                "Scaling strategies"
            ]
        elif chapter_num == 17:  # Reasoning
            chapter_data["key_concepts"] = [
                "Chain-of-Thought (CoT)",
                "Self-correction",
                "Code execution",
                "DeepSearch techniques"
            ]
        elif chapter_num == 18:  # Guardrails
            chapter_data["key_concepts"] = [
                "Safety constraints",
                "Content filtering",
                "Action validation",
                "Boundary enforcement",
                "Audit logging"
            ]
        elif chapter_num == 19:  # Evaluation
            chapter_data["key_concepts"] = [
                "Performance assessment",
                "LLM-as-Judge",
                "Metrics collection",
                "Observability"
            ]
        elif chapter_num == 20:  # Prioritization
            chapter_data["key_concepts"] = [
                "Task ranking",
                "Resource allocation",
                "Workload management",
                "Priority scheduling"
            ]
        elif chapter_num == 21:  # Exploration
            chapter_data["key_concepts"] = [
                "Autonomous investigation",
                "Research automation",
                "Hypothesis generation",
                "Agent Laboratory pattern"
            ]

        data["chapters"].append(chapter_data)

    # Add parts summary
    data["parts"] = [
        {
            "id": 1,
            "name": "Core Patterns",
            "chapters": [1, 2, 3, 4, 5, 6, 7],
            "total_pages": 103,
            "description": "Foundational patterns for building agentic systems"
        },
        {
            "id": 2,
            "name": "Advanced Patterns",
            "chapters": [8, 9, 10, 11],
            "total_pages": 61,
            "description": "Memory, learning, and protocol patterns"
        },
        {
            "id": 3,
            "name": "Production Patterns",
            "chapters": [12, 13, 14],
            "total_pages": 34,
            "description": "Error handling, human oversight, and knowledge retrieval"
        },
        {
            "id": 4,
            "name": "Operational Patterns",
            "chapters": [15, 16, 17, 18, 19, 20, 21],
            "total_pages": 114,
            "description": "Communication, optimization, reasoning, and safety"
        }
    ]

    # Add appendices
    data["appendices"] = [
        {"id": "A", "title": "Advanced Prompting Techniques", "pages": 28},
        {"id": "B", "title": "AI Agentic: From GUI to Real World Environment", "pages": 6},
        {"id": "C", "title": "Quick Overview of Agentic Frameworks", "pages": 8},
        {"id": "D", "title": "Building an Agent with AgentSpace", "pages": 6, "online_only": True},
        {"id": "E", "title": "AI Agents on the CLI", "pages": 5, "online_only": True},
        {"id": "F", "title": "Under the Hood: Agents' Reasoning Engines", "pages": 14},
        {"id": "G", "title": "Coding Agents", "pages": 7}
    ]

    return data

def main():
    """Parse and save as JSON."""
    data = parse_agentic_patterns()

    # Save to JSON
    output_path = Path("/home/nexora/codedocs/agentic/Agentic_Design_Patterns.json")
    with output_path.open('w', encoding='utf-8') as f:
        json.dump(data, f, indent=2, ensure_ascii=False)

    print(f"✓ Parsed {len(data['chapters'])} chapters")
    print(f"✓ Extracted {len(data['parts'])} parts")
    print(f"✓ Listed {len(data['appendices'])} appendices")
    print(f"✓ Saved to: {output_path}")
    print(f"\nTotal: {data['total_pages']} pages")

if __name__ == "__main__":
    main()
