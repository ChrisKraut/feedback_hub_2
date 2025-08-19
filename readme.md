# Feedback Hub 2

A modern feedback management system built with Go, featuring Domain-Driven Design (DDD) architecture, event-driven communication, and comprehensive role-based access control.

## 🏗️ Architecture Overview

Feedback Hub 2 follows strict Domain-Driven Design principles with clean domain boundaries and event-driven communication between domains.

### **DDD Architecture Principles**
- **Domain Isolation**: Each domain is completely isolated with no direct dependencies
- **Event-Driven Communication**: Cross-domain communication through shared event bus
- **Shared Layer Infrastructure**: Centralized persistence, auth, and web utilities
- **Clean Import Structure**: No cross-domain dependencies, proper DDD layering

## 📁 Project Structure

```
feedback_hub_2/
├── cmd/api/                    # Application entry point
├── internal/                   # Private application code
│   ├── shared/                # Shared code across all domains
│   │   ├── bus/              # Event bus and messaging
│   │   ├── persistence/      # Shared persistence utilities
│   │   ├── web/              # Shared web utilities
│   │   ├── auth/             # Shared authentication
│   │   ├── queries/          # Shared query services
│   │   └── bootstrap/        # System initialization
│   ├── user/                 # User domain module
│   │   ├── domain/          # User domain logic
│   │   ├── application/     # User application services
│   │   ├── infrastructure/  # User infrastructure
│   │   └── interfaces/      # User HTTP handlers
│   ├── role/                 # Role domain module
│   │   ├── domain/          # Role domain logic
│   │   ├── application/     # Role application services
│   │   ├── infrastructure/  # Role infrastructure
│   │   └── interfaces/      # Role HTTP handlers
│   └── idea/                 # Idea domain module
│       ├── domain/          # Idea domain logic
│       ├── application/     # Idea application services
│       ├── infrastructure/  # Idea infrastructure
│       └── interfaces/      # Idea HTTP handlers
├── docs/                     # Swagger documentation
├── pkg/                      # Public packages
├── scripts/                  # Database migrations
└── tests/                    # Integration tests
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 13+
- Docker (optional)

### Installation

1. **Clone the repository:**
```bash
git clone <repository-url>
cd feedback_hub_2
```

2. **Install dependencies:**
```bash
go mod tidy
```

3. **Set up database:**
```bash
# Run migrations
psql -U postgres -d feedback_hub -f scripts/migrate.sql
```

4. **Run the application:**
```bash
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`

## 🔐 Authentication & Authorization

### **Role-Based Access Control (RBAC)**

The system implements three main roles:

- **Super User**: Full system access, can create any user with any role
- **Product Owner**: Can create and manage contributors, limited role management
- **Contributor**: Basic access, can view and create ideas

### **JWT Authentication**

All API endpoints require JWT authentication. Include the token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

## 📚 API Documentation

Interactive API documentation is available via Swagger UI:
- **URL**: `http://localhost:8080/swagger/index.html`
- **Docs**: `http://localhost:8080/swagger/doc.json`

### **Key Endpoints**

#### **Authentication**
- `POST /auth/login` - User login
- `POST /auth/register` - User registration

#### **Users**
- `GET /users/{id}` - Get user by ID
- `POST /users` - Create new user
- `PUT /users/{id}` - Update user
- `DELETE /users/{id}` - Delete user

#### **Roles**
- `GET /roles` - Get all roles
- `GET /roles/{id}` - Get role by ID
- `POST /roles` - Create new role
- `PUT /roles/{id}` - Update role
- `DELETE /roles/{id}` - Delete role

#### **Ideas**
- `GET /ideas` - Get all ideas
- `GET /ideas/{id}` - Get idea by ID
- `POST /ideas` - Create new idea
- `PUT /ideas/{id}` - Update idea
- `DELETE /ideas/{id}` - Delete idea

## 🧪 Testing

### **Run All Tests**
```bash
go test ./...
```

### **Run Specific Test Suites**
```bash
# Integration tests
go test ./tests -v

# Domain tests
go test ./internal/user/domain -v
go test ./internal/role/domain -v
go test ./internal/idea/domain -v

# Shared layer tests
go test ./internal/shared/auth -v
go test ./internal/shared/bus -v
```

### **Test Coverage**
```bash
go test -cover ./...
```

## 🔧 Development

### **Adding New Features**

1. **Create Domain Model** (`internal/{domain}/domain/feature.go`)
2. **Create Repository Interface** (`internal/{domain}/domain/repository.go`)
3. **Implement Repository** (`internal/{domain}/infrastructure/feature_repository.go`)
4. **Create Application Service** (`internal/{domain}/application/feature_service.go`)
5. **Create HTTP Handler** (`internal/{domain}/interfaces/feature_handler.go`)
6. **Add Tests** for each layer
7. **Update Swagger Documentation**

### **Code Organization Principles**

- **Domain Layer**: Pure business logic, no external dependencies
- **Application Layer**: Orchestrates domain entities and repositories
- **Infrastructure Layer**: Implements interfaces defined in domain
- **Interface Layer**: HTTP transport, validation, and error handling

### **AI-Friendly Development**

- Add comprehensive comments starting with `// AI-hint:` for future AI iterations
- Document business rules and domain invariants
- Explain architectural decisions and patterns

## 📊 Database Schema

The system uses PostgreSQL with the following main tables:

- **users**: User accounts and authentication
- **roles**: System roles and permissions
- **ideas**: Feedback ideas and suggestions
- **user_roles**: User-role assignments

Run `scripts/migrate.sql` to set up the database schema.

## 🚀 Deployment

### **Docker Deployment**
```bash
# Build image
docker build -t feedback-hub-2 .

# Run container
docker run -p 8080:8080 feedback-hub-2
```

### **Environment Variables**
- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_NAME`: Database name
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password
- `JWT_SECRET`: JWT signing secret

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### **Development Guidelines**

- Follow DDD principles strictly
- Maintain domain isolation
- Use events for cross-domain communication
- Add comprehensive tests
- Update documentation for new features

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For questions and support:
- Check the [Development Guide](DEVELOPMENT_GUIDE.md)
- Review the [DDD Refactoring Plan](DDD_REFACTORING_PLAN.md)
- Open an issue on GitHub

## 🎯 Roadmap

- [ ] Enhanced analytics and reporting
- [ ] Real-time notifications
- [ ] Advanced search and filtering
- [ ] API rate limiting
- [ ] Multi-tenant support
- [ ] Mobile application

---

**Built with ❤️ using Go and Domain-Driven Design principles**