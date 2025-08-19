---
description: AI Development Rules: Go DDD Project
PREAMBLE: This document contains the complete and mandatory rules for code generation. Adhere to these rules without exception.

RULE 1: DIRECTORY STRUCTURE
1.1. The root of the domain-specific code is internal/.

1.2. The primary organization unit is the domain module (e.g., internal/user/, internal/idea/).

1.3. Each domain module MUST contain exactly four subdirectories:

domain/

application/

infrastructure/

interfaces/

1.4. A shared/ directory exists at internal/shared/. This directory is for code that is generic and shared across all domains.

shared/ MUST NOT contain any business logic specific to any domain.

shared/ contains subdirectories like bus/, persistence/, web/.

RULE 2: DEPENDENCY FLOW
2.1. The dependency direction is absolute and MUST be followed:

interfaces depends on application.

application depends on domain.

infrastructure depends on domain.

2.2. The domain layer MUST NOT depend on any other layer within its module. It is the core.

2.3. Layers can only depend on layers immediately inward to them as specified in Rule 2.1. No layer skipping is permitted.

RULE 3: LAYER DEFINITIONS AND CONSTRAINTS
3.1. domain Layer
PURPOSE: Core business logic, rules, and models.

MUST CONTAIN:

Entities, Aggregates, and Value Objects.

Repository interfaces (e.g., type UserRepository interface).

Domain Services.

Domain Events.

MUST NOT CONTAIN:

Any code from external libraries (no database drivers, web frameworks, API clients).

Implementations of repositories.

Any knowledge of the outside world (HTTP, SQL, etc.).

CONSTRAINT: Must be 100% standard library Go.

3.2. application Layer
PURPOSE: Orchestrate domain logic to execute specific use cases.

MUST CONTAIN:

Application Services (e.g., UserService). Each public method is a distinct use case.

Data Transfer Objects (DTOs) for input and output.

MUST NOT CONTAIN:

Business rule logic. All business rules are delegated to domain objects.

CONSTRAINT: Only depends on the domain layer.

3.3. infrastructure Layer
PURPOSE: Provide concrete implementations of interfaces defined in the domain layer.

MUST CONTAIN:

Repository implementations (e.g., PostgresUserRepository that implements domain.UserRepository).

Clients for external services (e.g., payment gateways, email services).

Concrete event bus implementations.

CONSTRAINT: Implements interfaces from the domain layer. Can use any external library required.

3.4. interfaces Layer
PURPOSE: Entry point for external actors (users, other systems).

MUST CONTAIN:

API handlers (REST, gRPC).

Command-Line Interface (CLI) command definitions.

MUST NOT CONTAIN:

Business logic.

Application orchestration logic.

FUNCTION:

Receive external request.

Validate and deserialize input into an application layer DTO.

Call a single method on an application service.

Serialize the output DTO into a response and send it.

CONSTRAINT: Only depends on the application layer.

RULE 4: INTER-DOMAIN COMMUNICATION
4.1. Direct communication between domain modules is FORBIDDEN. A service in the user domain cannot import or call a service in the idea domain.

4.2. All inter-domain communication MUST be asynchronous using a shared Event Bus.

4.3. MECHANISM:

A domain's application service publishes a domain event after a state change (e.g., UserRegistered).

Event structs are defined in the shared/ directory (e.g., shared/events/user.go).

The publishing domain has zero knowledge of subscribers.

Subscribing domains create handlers (typically in their interfaces layer) that listen for events from the bus and trigger corresponding use cases in their own application layer.

RULE 5: CODE GENERATION MANDATES
5.1. Dependency Injection: All dependencies (repositories, services, buses) MUST be injected via struct constructors. DO NOT use global variables or singletons for dependencies.

5.2. File Granularity: Each file must have a single, clear purpose (e.g., one file per repository implementation, one file per aggregate).

5.3. Explicitness: Use clear and descriptive names for files, types, and functions (e.g., CreateUserRequest, PostgresUserRepository).

5.4. Encapsulation: Keep types, fields, and functions private (lowercase) by default. Only export (uppercase) what is explicitly required by an outer layer.

RULE 6: PROJECT OUTLINE & SETUP
6.1. Standard Project Layout:

/
├── cmd/
│   └── server/
│       └── main.go         # Main application entry point. Wires everything together.
├── internal/
│   ├── user/               # Domain module
│   ├── idea/               # Domain module
│   └── shared/
├── go.mod
├── go.sum
└── .env                    # Environment variables

6.2. Main Entry Point (cmd/server/main.go):
This file is the ONLY place where the entire application is assembled. Its responsibility is to perform dependency injection (wiring).

Execution Order:

Load Configuration: Load settings from environment variables (.env).

Establish Connections: Initialize database connections, message queues, etc.

Initialize Shared Services: Instantiate the shared Event Bus.

Instantiate Domain by Domain (Bottom-Up): For each domain (e.g., user):
a. Instantiate infrastructure layer components (e.g., NewPostgresUserRepository(db)).
b. Instantiate application layer services, injecting the infrastructure components (e.g., NewUserService(userRepo, eventBus)).
c. Instantiate interfaces layer handlers, injecting the application services (e.g., NewUserHandler(userService)).

Setup Router: Initialize the HTTP router (e.g., chi, gin) and register the handlers from all domains.

Register Event Subscribers: Subscribe event handlers from various domains to the event bus.

Start Server: Start the HTTP server.