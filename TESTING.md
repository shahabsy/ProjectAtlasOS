# Atlas OS - Database Testing Guide

This guide validates graph.db SQLite database.

## Quick Validation

sqlite3 .atlas/graph.db \"SELECT COUNT(*) FROM nodes;\"

## Test Files

1. test_queries.sql - Full query suite
2. complex_queries.sql - Advanced scenarios  
3. test_db.sql - Quick health check

## Integration Testing

go test ./test/... -v
## Validation Commands

.tables
SELECT type, COUNT(*) FROM nodes GROUP BY type;
