# Go Zakat - Zakat Management System API

RESTful API untuk sistem manajemen zakat yang dibangun dengan Go, Gin, dan PostgreSQL.

## 🚀 Tech Stack

- **Go** 1.25
- **Gin** - HTTP web framework
- **PostgreSQL** - Database
- **JWT** - Authentication
- **Google OAuth2** - Social login
- **Swagger** - API documentation
- **golang-migrate** - Database migrations

## 📋 Features

### ✅ Implemented

#### Authentication & Authorization
- User registration & login (email/password)
- Google OAuth2 login (web & mobile)
- JWT-based authentication (Access Token + Refresh Token)
- Token refresh mechanism
- Role-based access control (admin, operator, user)

#### Muzakki Management (Pemberi Zakat)
- CRUD operations
- Search by name or phone number
- Pagination support
- Unique phone number validation

#### Asnaf Management (Kategori Penerima Zakat)
- CRUD operations
- Search by name
- Pagination support
- 8 kategori asnaf sesuai syariat

#### Mustahiq Management (Penerima Zakat)
- CRUD operations
- Search by name or address
- Filter by status (active, inactive, pending)
- Filter by asnaf category
- Nested asnaf info in response
- Status management with constants
- Pagination support

### 📝 TODO

- [ ] **Program Management** - Manajemen program zakat
- [ ] **Donation Management** - Pencatatan donasi
- [ ] **Donation Receipt** - Kwitansi donasi
- [ ] **Donation Receipt Items** - Detail item donasi
- [ ] **Distribution Management** - Penyaluran zakat
- [ ] **Reports & Analytics** - Laporan dan statistik
- [ ] **Notifications** - Email/SMS notifications

## 🛠️ Setup

### Prerequisites

- Go 1.25+
- PostgreSQL 14+
- Google OAuth2 credentials (optional, for social login)

### Installation

1. **Clone repository**
   ```bash
   git clone <repository-url>
   cd go-zakat
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Setup environment variables**
   ```bash
   cp .env_example .env
   ```
   
   Edit `.env` file:
   ```env
   # Server
   PORT=8080
   
   # Database
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=your_password
   DB_NAME=go_zakat
   
   # JWT
   JWT_SECRET=your-secret-key
   JWT_ACCESS_EXPIRY=15m
   JWT_REFRESH_EXPIRY=168h
   
   # Google OAuth (optional)
   GOOGLE_CLIENT_ID=your-client-id
   GOOGLE_CLIENT_SECRET=your-client-secret
   GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback
   ```

4. **Run database migrations**
   ```bash
   # Install golang-migrate
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   
   # Run migrations
   migrate -path migrations -database "postgresql://user:password@localhost:5432/go_zakat?sslmode=disable" up
   ```

5. **Run the application**
   ```bash
   go run cmd/api/main.go
   ```

   Server akan berjalan di `http://localhost:8080`

## 📚 API Documentation

Swagger documentation tersedia di: `http://localhost:8080/swagger/index.html`

### Generate Swagger Docs

```bash
# Install swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g cmd/api/main.go
```

## 🔑 API Endpoints

### Authentication
```
POST   /api/v1/auth/register              - Register new user
POST   /api/v1/auth/login                 - Login with email/password
POST   /api/v1/auth/refresh               - Refresh access token
GET    /api/v1/auth/me                    - Get current user info
GET    /api/v1/auth/google/login          - Google OAuth login (web)
GET    /api/v1/auth/google/callback       - Google OAuth callback
POST   /api/v1/auth/google/mobile/login   - Google OAuth login (mobile)
```

### Muzakki (Protected)
```
GET    /api/v1/muzakki                    - Get all muzakki (with search & pagination)
GET    /api/v1/muzakki/:id                - Get muzakki by ID
POST   /api/v1/muzakki                    - Create new muzakki
PUT    /api/v1/muzakki/:id                - Update muzakki
DELETE /api/v1/muzakki/:id                - Delete muzakki
```

### Asnaf (Protected)
```
GET    /api/v1/asnaf                      - Get all asnaf (with search & pagination)
GET    /api/v1/asnaf/:id                  - Get asnaf by ID
POST   /api/v1/asnaf                      - Create new asnaf
PUT    /api/v1/asnaf/:id                  - Update asnaf
DELETE /api/v1/asnaf/:id                  - Delete asnaf
```

### Mustahiq (Protected)
```
GET    /api/v1/mustahiq                   - Get all mustahiq (with filters & pagination)
GET    /api/v1/mustahiq/:id               - Get mustahiq by ID
POST   /api/v1/mustahiq                   - Create new mustahiq
PUT    /api/v1/mustahiq/:id               - Update mustahiq
DELETE /api/v1/mustahiq/:id               - Delete mustahiq
```

**Query Parameters for GET all:**
- `q` - Search by name/address
- `status` - Filter by status (active, inactive, pending)
- `asnafID` - Filter by asnaf category
- `page` - Page number (default: 1)
- `per_page` - Items per page (default: 10)

## 🏗️ Project Structure

```
go-zakat/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── delivery/
│   │   └── http/
│   │       ├── dto/                # Data Transfer Objects
│   │       ├── handler/            # HTTP handlers
│   │       └── middleware/         # Middleware (auth, cors, etc)
│   ├── domain/
│   │   ├── entity/                 # Domain entities
│   │   └── repository/             # Repository interfaces
│   ├── infrastructure/
│   │   ├── database/               # Database connection
│   │   ├── oauth/                  # OAuth state management
│   │   └── service/                # External services (Google, JWT)
│   ├── repository/
│   │   └── postgres/               # PostgreSQL implementations
│   └── usecase/                    # Business logic
├── migrations/                     # Database migrations
├── pkg/
│   ├── config/                     # Config implementations
│   ├── database/                   # Database implementations
│   ├── logger/                     # Logger implementations
│   └── response/                   # Standardized API responses
├── docs/                           # Swagger documentation
├── .env                            # Environment variables
└── go.mod                          # Go dependencies
```

## 🔐 Authentication Flow

1. **Register/Login** → Receive Access Token (15 min) + Refresh Token (7 days)
2. **API Requests** → Include `Authorization: Bearer <access_token>` header
3. **Token Expired** → Use `/api/v1/auth/refresh` with Refresh Token
4. **Refresh Token Expired** → Login again

## 🗄️ Database Schema

### Users
- Authentication & user management
- Roles: admin, operator, user

### Muzakki
- Pemberi zakat (donors)
- Fields: name, phoneNumber, address, notes

### Asnaf
- Kategori penerima zakat (8 categories)
- Fields: name, description

### Mustahiq
- Penerima zakat (beneficiaries)
- Fields: name, phoneNumber, address, asnafID, status, description
- Status: active, inactive, pending (default: pending)
- Foreign key to Asnaf

## 📝 Notes

- All protected endpoints require valid JWT token
- Default status for new Mustahiq is `pending`
- Use constants from `entity.MustahiqStatus*` for status values
- Google OAuth state is stored in-memory (consider Redis for production)
- Phone numbers must be unique for Muzakki and Mustahiq

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License.

## 👥 Authors

- Muhammad Dila

## 🙏 Acknowledgments

- Gin framework team
- PostgreSQL community