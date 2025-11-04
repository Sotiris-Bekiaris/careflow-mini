# Scripts Directory

This directory contains utility scripts for development, testing, and deployment.

## Directory Structure

```
scripts/
├── db/              # Database-related scripts
│   ├── init.sql     # Database schema initialization
│   └── seed.sql     # Sample data seeding
├── demo/            # Demo and testing scripts
│   └── run-demo.sh  # Interactive demo script
└── setup.sh         # Development environment setup
```

## Scripts

### setup.sh
Sets up the local development environment:
- Checks prerequisites (Go, Docker, etc.)
- Installs Go dependencies
- Starts infrastructure services (Docker Compose)
- Initializes database

**Usage:**
```bash
./scripts/setup.sh
```

### db/init.sql
Creates database schema:
- Tables for patients, appointments, observations
- Indexes for query optimization
- Triggers for updated_at timestamps

**Usage:**
```bash
docker exec -i careflow-postgres psql -U careflow -d careflow < scripts/db/init.sql
```

### db/seed.sql
Populates database with sample data:
- Sample patients
- Sample appointments
- Sample lab observations

**Usage:**
```bash
docker exec -i careflow-postgres psql -U careflow -d careflow < scripts/db/seed.sql
```

### demo/run-demo.sh
Runs an interactive demo showcasing the system:
- Creates patients
- Schedules appointments
- Submits HL7 lab results
- Shows observability features

**Usage:**
```bash
./scripts/demo/run-demo.sh
```

## Making Scripts Executable

After cloning the repository, make scripts executable:

```bash
chmod +x scripts/*.sh
chmod +x scripts/demo/*.sh
```

## Adding New Scripts

When adding new scripts:
1. Place them in the appropriate subdirectory
2. Make them executable: `chmod +x script-name.sh`
3. Add usage documentation to this README
4. Include error handling (`set -e`)
5. Add helpful echo messages
