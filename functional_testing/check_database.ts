import { DB } from './db' // this is the Database interface we defined earlier
import { Pool } from 'pg'
import { Kysely, PostgresDialect } from 'kysely'
import fs from 'node:fs'

const dialect = new PostgresDialect({
  pool: new Pool({
    database: 'webdrones',
    host: 'localhost',
    user: 'user',
    port: 5432,
    max: 10,
    password: fs.readFileSync("../postgres_pw.txt", "ascii")
  })
})

// Database interface is passed to Kysely's constructor, and from now on, Kysely 
// knows your database structure.
// Dialect is passed to Kysely's constructor, and from now on, Kysely knows how 
// to communicate with your database.
export const db = new Kysely<DB>({
  dialect,
});
(async () => 
    {
        const tables = await db.introspection.getTables();
        for (const table of tables.map(t=>t.name).filter(name => name !== 'http_sessions')){
            console.log(`Checking ${table} for unset create and update columns.`)
            const records = await db.selectFrom(table as keyof DB).select(['created_at', 'updated_at']).execute();
            if (records.filter(r=>r.created_at === null || r.updated_at === null).length > 0) {
                console.log(`${table} has nulls`);
            }
        }
    }
)();