# Go Zakat - Zakat Management System API

RESTful API untuk sistem manajemen Zakat, Infaq, dan Sadaqah (ZIS) yang dibangun dengan Go, Gin, dan PostgreSQL.

## 🚀 Tech Stack

- **Go** 1.25
- **Gin** - HTTP web framework
- **PostgreSQL** - Database with pgx driver
- **JWT** - Authentication (Access + Refresh Token)
- **Google OAuth2** - Social login (Web & Mobile)
- **Swagger** - API documentation
- **golang-migrate** - Database migrations
- **go-playground/validator** - Input validation

## 📋 Features

### ✅ Implemented

#### 🔐 Authentication & Authorization
- User registration & login (email/password)
- Google OAuth2 login (web & mobile)
- JWT-based authentication (Access Token 15m + Refresh Token 7d)
- Token refresh mechanism
- Role-based access control (admin, operator, user)
- Protected routes with middleware

#### 👥 Master Data Management

**Muzakki (Pemberi Zakat/Donors)**
- Full CRUD operations
- Search by name or phone number
- Pagination support
- Unique phone number validation

**Asnaf (8 Golongan Penerima Zakat)**
- Full CRUD operations
- Search by name
- Pagination support
- 8 kategori sesuai syariat Islam:
  - Fakir, Miskin, Amil, Muallaf, Riqab, Gharimin, Fisabilillah, Ibnu Sabil

**Mustahiq (Penerima Zakat/Beneficiaries)**
- Full CRUD operations
- Search by name or address
- Filter by status (active, inactive, pending)
- Filter by asnaf category
- Nested asnaf info in response
- Status management with constants

**Program (Program Penyaluran)**
- Full CRUD operations
- Search by name
- Filter by type (zakat, infaq, sadaqah, umum)
- Filter by active status
- Pagination support

#### 💰 Transaction Management

**Donation Receipts (Penerimaan Dana)**
- Full CRUD with nested items (header-detail pattern)
- Auto-generate receipt number
- Support multiple fund types: zakat (fitrah/maal), infaq, sadaqah
- Zakat fitrah: person count & rice (kg) tracking
- Complex filtering: date range, fund type, zakat type, payment method, muzakki
- Search in muzakki name or notes
- Transaction-based create/update for data integrity
- Audit trail (created_by_user_id from JWT)

**Distributions (Penyaluran Dana)**
- Full CRUD with nested items (header-detail pattern)
- Link to programs (optional)
- Support 4 source fund types: zakat_fitrah, zakat_maal, infaq, sadaqah
- Multiple mustahiq per distribution
- Complex filtering: date range, source fund type, program
- Search in program name or notes
- Beneficiary count calculation
- Transaction-based create/update
- Audit trail (created_by_user_id from JWT)

#### 📊 Reports & Analytics

**Income Summary (Penghimpunan)**
- Group by daily or monthly
- Breakdown by fund type (zakat_fitrah, zakat_maal, infaq, sadaqah)
- Date range filtering
- CASE WHEN pivoting for fund types

**Distribution Summary (Penyaluran)**
- Group by asnaf or program
- Beneficiary count (COUNT DISTINCT)
- Total amount per group
- Filter by source fund type
- Date range filtering

**Fund Balance (Saldo Dana)**
- Total in vs total out per fund type
- Balance calculation
- CTE-based query for performance
- Fund type mapping from donation receipts

**Mustahiq History**
- Distribution history per mustahiq
- Total received calculation
- Nested program and asnaf info

### 🎯 API Design Principles

- ✅ Standardized JSON response format (`success`, `message`, `data`)
- ✅ Consistent pagination (`items` + `meta`)
- ✅ Comprehensive error handling
- ✅ Input validation with detailed error messages
- ✅ RESTful conventions
- ✅ Swagger documentation for all endpoints

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

**Query Parameters:**
- `q` - Search by name/address
- `status` - Filter by status (active, inactive, pending)
- `asnafID` - Filter by asnaf category
- `page` - Page number (default: 1)
- `per_page` - Items per page (default: 10)

### Programs (Protected)
```
GET    /api/v1/programs                   - Get all programs (with filters & pagination)
GET    /api/v1/programs/:id               - Get program by ID
POST   /api/v1/programs                   - Create new program
PUT    /api/v1/programs/:id               - Update program
DELETE /api/v1/programs/:id               - Delete program
```

**Query Parameters:**
- `q` - Search by name
- `type` - Filter by type (zakat, infaq, sadaqah, umum)
- `active` - Filter by active status (true, false)
- `page`, `per_page` - Pagination

### Donation Receipts (Protected)
```
GET    /api/v1/donation-receipts          - Get all receipts (with filters & pagination)
GET    /api/v1/donation-receipts/:id      - Get receipt by ID (with items)
POST   /api/v1/donation-receipts          - Create new receipt with items
PUT    /api/v1/donation-receipts/:id      - Update receipt with items
DELETE /api/v1/donation-receipts/:id      - Delete receipt (cascade items)
```

**Query Parameters:**
- `date_from`, `date_to` - Date range filter (YYYY-MM-DD)
- `fund_type` - Filter by fund type (zakat, infaq, sadaqah)
- `zakat_type` - Filter by zakat type (fitrah, maal)
- `payment_method` - Filter by payment method
- `muzakki_id` - Filter by muzakki
- `q` - Search in muzakki name or notes
- `page`, `per_page` - Pagination

### Distributions (Protected)
```
GET    /api/v1/distributions              - Get all distributions (with filters & pagination)
GET    /api/v1/distributions/:id          - Get distribution by ID (with items)
POST   /api/v1/distributions              - Create new distribution with items
PUT    /api/v1/distributions/:id          - Update distribution with items
DELETE /api/v1/distributions/:id          - Delete distribution (cascade items)
```

**Query Parameters:**
- `date_from`, `date_to` - Date range filter (YYYY-MM-DD)
- `source_fund_type` - Filter by source fund type
- `program_id` - Filter by program
- `q` - Search in program name or notes
- `page`, `per_page` - Pagination

### Reports (Protected, Read-only)
```
GET    /api/v1/reports/income-summary           - Income summary report
GET    /api/v1/reports/distribution-summary     - Distribution summary report
GET    /api/v1/reports/fund-balance             - Fund balance report
GET    /api/v1/reports/mustahiq-history/:id     - Mustahiq distribution history
```

**Income Summary Query Parameters:**
- `date_from`, `date_to` - Date range
- `group_by` - daily or monthly (default: monthly)

**Distribution Summary Query Parameters:**
- `date_from`, `date_to` - Date range
- `group_by` - asnaf or program (required)
- `source_fund_type` - Filter by fund type (optional)

**Fund Balance Query Parameters:**
- `date_from`, `date_to` - Date range (optional)

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
├── migrations/                     # Database migrations (9 migrations)
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

### Core Tables

**users** - Authentication & user management
- Roles: admin, operator, user
- OAuth support (Google)

**muzakki** - Pemberi zakat (donors)
- Unique phone number
- Address, notes

**asnaf** - 8 Golongan penerima zakat
- Fakir, Miskin, Amil, Muallaf, Riqab, Gharimin, Fisabilillah, Ibnu Sabil

**mustahiq** - Penerima zakat (beneficiaries)
- Foreign key to asnaf
- Status: active, inactive, pending (default: pending)

**programs** - Program penyaluran
- Type: zakat, infaq, sadaqah, umum
- Active status flag

### Transaction Tables

**donation_receipts** - Header penerimaan dana
- Foreign key to muzakki
- Foreign key to users (created_by)
- Unique receipt number
- Payment method tracking

**donation_receipt_items** - Detail penerimaan dana
- Foreign key to donation_receipts (CASCADE delete)
- Fund type: zakat, infaq, sadaqah
- Zakat type: fitrah, maal (for zakat only)
- Person count & rice kg (for zakat fitrah)

**distributions** - Header penyaluran dana
- Foreign key to programs (optional, RESTRICT delete)
- Foreign key to users (created_by)
- Source fund type: zakat_fitrah, zakat_maal, infaq, sadaqah

**distribution_items** - Detail penyaluran dana
- Foreign key to distributions (CASCADE delete)
- Foreign key to mustahiq (RESTRICT delete)

### Relationships

```
muzakki ←─── donation_receipts ←─── donation_receipt_items
                                           ↓
                                      (fund mapping)
                                           ↓
programs ←─── distributions ←─── distribution_items ───→ mustahiq ───→ asnaf
   ↑                                                          ↑
   └──────────────────────────────────────────────────────────┘
                    (reports group by asnaf/program)
```

## 📊 Business Logic

### Fund Type Mapping

**Donation Receipts** (flexible format):
- `fund_type = "zakat"` + `zakat_type = "fitrah"` → zakat_fitrah
- `fund_type = "zakat"` + `zakat_type = "maal"` → zakat_maal
- `fund_type = "infaq"` → infaq
- `fund_type = "sadaqah"` → sadaqah

**Distributions** (standard format):
- `source_fund_type` directly uses: zakat_fitrah, zakat_maal, infaq, sadaqah

### Validation Rules

**Donation Receipts:**
- Zakat must have zakat_type (fitrah/maal)
- Zakat fitrah must have person_count
- All amounts must be > 0
- Total amount auto-calculated from items
- Receipt number must be unique
- Muzakki must exist

**Distributions:**
- Items array must have at least 1 item
- All amounts must be > 0
- Total amount auto-calculated from items
- All mustahiq must exist
- Source fund type must be valid

## 📝 Notes

- All protected endpoints require valid JWT token
- Default status for new Mustahiq is `pending`
- Google OAuth state is stored in-memory (consider Redis for production)
- Phone numbers must be unique for Muzakki and Mustahiq
- Receipt numbers are auto-generated and unique
- All create/update operations for receipts and distributions use database transactions
- Audit trail: `created_by_user_id` automatically captured from JWT token
- Date fields in database are DATE type, converted to YYYY-MM-DD string in API responses

## 🐛 Known Issues

- Reports API may need optimization for large datasets (consider adding indexes or materialized views)
- Some complex SQL queries in reports may need performance tuning

## 🚀 Future Enhancements

- [ ] Notifications (Email/SMS)
- [ ] Export reports to PDF/Excel
- [ ] Dashboard with charts
- [ ] Batch import/export (CSV)
- [ ] Redis caching for reports
- [ ] Audit log for all transactions
- [ ] Multi-language support

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
- Swaggo team for Swagger integration