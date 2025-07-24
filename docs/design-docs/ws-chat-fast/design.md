# WebSocket Chat (FastAPI)

## Context

Real-time chat application demonstrating WebSocket implementation with FastAPI.

## Problem & Goals

- Build low-latency chat system
- Demonstrate WebSocket handling
- Show auto-reconnection patterns
- Include basic metrics

## Constraints & Risks

- Single server deployment
- In-memory state (Redis for scaling)
- Basic authentication

## Architecture & Alternatives

- FastAPI WebSocket endpoints
- Redis for message persistence
- Simple HTML/JS frontend

## Trade-offs

- Simplicity over scalability
- Development speed over production features

## Results & Metrics

- Sub-100ms message latency
- Auto-reconnection on disconnect
- Basic connection metrics

## What I'd change next time

- Add proper authentication
- Implement message persistence
- Add horizontal scaling
