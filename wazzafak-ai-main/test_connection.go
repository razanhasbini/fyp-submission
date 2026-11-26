package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env if it exists
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("❌ DATABASE_URL environment variable is not set")
	}

	fmt.Println("🔍 Supabase Connection Diagnostic Tool")
	fmt.Println("=====================================")
	fmt.Println()

	// Log connection details (without password)
	dsnForLog := dsn
	if strings.Contains(dsnForLog, "@") {
		parts := strings.Split(dsnForLog, "@")
		if len(parts) == 2 {
			if strings.Contains(parts[0], ":") {
				userPass := strings.Split(parts[0], ":")
				if len(userPass) >= 2 {
					dsnForLog = userPass[0] + ":***@" + parts[1]
				}
			}
		}
	}
	fmt.Printf("📋 Connection String: %s\n", dsnForLog)
	fmt.Println()

	// Check connection string format
	fmt.Println("1️⃣ Checking connection string format...")
	if !strings.Contains(dsn, "postgres") {
		fmt.Println("   ⚠️  Warning: Connection string doesn't contain 'postgres'")
	}
	if strings.Contains(dsn, ":6543") {
		fmt.Println("   ✅ Using pooler connection (port 6543)")
	} else if strings.Contains(dsn, ":5432") {
		fmt.Println("   ✅ Using direct connection (port 5432)")
	} else {
		fmt.Println("   ⚠️  Warning: No recognized port found")
	}
	if strings.Contains(dsn, "prepareThreshold=0") {
		fmt.Println("   ✅ Prepared statements disabled (prepareThreshold=0)")
	} else {
		fmt.Println("   ⚠️  Warning: prepareThreshold=0 not found - may cause issues")
	}
	if strings.Contains(dsn, "sslmode=require") {
		fmt.Println("   ✅ SSL mode set to require")
	} else {
		fmt.Println("   ⚠️  Warning: sslmode=require not found")
	}
	fmt.Println()

	// Add prepareThreshold if missing
	if !strings.Contains(dsn, "prepareThreshold") {
		if strings.Contains(dsn, "?") {
			dsn += "&prepareThreshold=0"
		} else {
			dsn += "?prepareThreshold=0"
		}
		fmt.Println("   ✅ Added prepareThreshold=0 to connection string")
	}

	// Add sslmode if missing
	if !strings.Contains(dsn, "sslmode") {
		if strings.Contains(dsn, "?") {
			dsn += "&sslmode=require"
		} else {
			dsn += "?sslmode=require"
		}
		fmt.Println("   ✅ Added sslmode=require to connection string")
	}

	// Add statement_cache_mode if missing
	if !strings.Contains(dsn, "statement_cache_mode") {
		dsn += "&statement_cache_mode=describe"
		fmt.Println("   ✅ Added statement_cache_mode=describe to connection string")
	}
	fmt.Println()

	// Test connection
	fmt.Println("2️⃣ Testing database connection...")
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		fmt.Printf("   ❌ Failed to open connection: %v\n", err)
		fmt.Println()
		fmt.Println("💡 Possible causes:")
		fmt.Println("   - Invalid connection string format")
		fmt.Println("   - Network connectivity issues")
		fmt.Println("   - Driver not installed")
		os.Exit(1)
	}
	defer db.Close()

	// Set connection pool settings
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test ping with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("   🔄 Attempting to ping database (30s timeout)...")
	if err := db.PingContext(ctx); err != nil {
		fmt.Printf("   ❌ Connection failed: %v\n", err)
		fmt.Println()
		fmt.Println("🔍 TROUBLESHOOTING:")
		fmt.Println()
		fmt.Println("1. Check Supabase Dashboard:")
		fmt.Println("   - Go to: https://app.supabase.com")
		fmt.Println("   - Verify project 'npeusanizvcyjwsgbhfn' is active (not paused)")
		fmt.Println("   - Check Settings > Database > Connection string")
		fmt.Println()
		fmt.Println("2. Verify Password:")
		fmt.Println("   - Get password from: Supabase Dashboard > Settings > Database")
		fmt.Println("   - URL-encode special characters (? → %3F, @ → %40)")
		fmt.Println("   - Example: '[YOUR_SUPABASE_DB_PASSWORD]' becomes '[YOUR_SUPABASE_DB_PASSWORD]'")
		fmt.Println()
		fmt.Println("3. Check Network/Firewall:")
		fmt.Println("   - Ensure port 6543 (pooler) or 5432 (direct) is not blocked")
		fmt.Println("   - Try direct connection if pooler fails")
		fmt.Println()
		fmt.Println("4. Test in Supabase SQL Editor:")
		fmt.Println("   - Try running a query in Supabase Dashboard > SQL Editor")
		fmt.Println("   - If that works, the issue is with the connection string")
		fmt.Println()
		fmt.Println("5. Common Issues:")
		fmt.Println("   - Project paused: Check Supabase dashboard")
		fmt.Println("   - Wrong password: Reset in Supabase dashboard")
		fmt.Println("   - Wrong region: Check connection string region matches project")
		fmt.Println("   - IP restrictions: Check Supabase network settings")
		os.Exit(1)
	}
	fmt.Println("   ✅ Ping successful!")
	fmt.Println()

	// Test query
	fmt.Println("3️⃣ Testing database query...")
	testCtx, testCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer testCancel()

	var version string
	if err := db.QueryRowContext(testCtx, "SELECT version()").Scan(&version); err != nil {
		fmt.Printf("   ⚠️  Query test failed: %v\n", err)
		fmt.Println("   💡 Connection works but queries may have issues")
	} else {
		fmt.Printf("   ✅ Query successful!\n")
		fmt.Printf("   📊 Database version: %s\n", version)
	}
	fmt.Println()

	// Test schema access
	fmt.Println("4️⃣ Testing schema access...")
	schemas := []string{"ai", "public"}
	for _, schema := range schemas {
		var exists bool
		checkQuery := fmt.Sprintf(`
			SELECT EXISTS(
				SELECT 1 FROM information_schema.schemata 
				WHERE schema_name = $1
			)
		`)
		if err := db.QueryRowContext(testCtx, checkQuery, schema).Scan(&exists); err != nil {
			fmt.Printf("   ⚠️  Could not check schema '%s': %v\n", schema, err)
		} else if exists {
			fmt.Printf("   ✅ Schema '%s' exists\n", schema)
		} else {
			fmt.Printf("   ⚠️  Schema '%s' does not exist\n", schema)
		}
	}
	fmt.Println()

	fmt.Println("✅ Connection diagnostic complete!")
	fmt.Println()
	fmt.Println("💡 If all tests passed, your connection is working correctly.")
	fmt.Println("   If issues persist, check the troubleshooting steps above.")
}

