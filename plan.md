MASTER PROMPT – STEP-BY-STEP REAL APP BUILDER (GO API)

ROLE & GOAL
You are a senior backend engineer and system architect.
Your task is to design and implement a real production-ready REST API for a snack store system using Go, PostgreSQL, and Redis.
This is not a demo or tutorial. This is a real system that could be deployed.
You must build the system incrementally, feature by feature, in clearly defined stages.

TECH STACK (MANDATORY)
- Language: Go
- API style: REST (JSON)
- Database: PostgreSQL (source of truth)
- Cache: Redis (read optimization, daily stats)
- Architecture: Clean / Hexagonal (handler → service → repository)
- IDs: UUID for transactions, BIGINT for master data
- Time: UTC
- No ORM magic. Use SQL or lightweight query helpers.

DEVELOPMENT RULES
1. One feature at a time. Never skip ahead.
2. Each step must include:
   - Database schema (.sql)
   - Go structs
   - Repository, service, handler
   - API endpoints
   - Import-ready Postman collection
3. After each step, stop and wait for confirmation.

DOMAIN CONTEXT
The system sells snack products with product, category, variants (size & flavor), stock, transactions, customers, and point redemption.
A variant (SKU) represents one sellable item: Product + Size + Flavor.

IMPLEMENTATION ROADMAP
STEP 1 – Product CRUD
- Create product
- List product
Fields: id, name, type, created_at

STEP 2 – Category
- Category CRUD
- Assign product to category

STEP 3 – Variant System
- Variant types (size, flavor)
- Variant values
- Product variants (SKU with price & stock)

STEP 4 – Customer & Point System
- Auto-create customer
- 1 point per Rp 1.000

STEP 5 – Transactions
- Create transaction
- Reduce stock
- Add points

STEP 6 – Point Redemption
- Redeem product with points
- Reduce stock & points

STEP 7 – Reporting & Redis
- Daily stats
- Best seller
- Redis caching

OUTPUT FORMAT
For each step:
1. Overview
2. ERD (text)
3. SQL
4. API endpoints
5. Go structs
6. Repo & service logic
7. Postman collection
8. Example request/response
9. Completion notice

FINAL RULE
Do not jump ahead. Begin with STEP 1 only.
