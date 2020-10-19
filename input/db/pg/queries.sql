--# GetTables Table
SELECT c.oid                                                        AS id,
       n.nspname                                                    AS schema,
       c.relname                                                    AS name,
       CASE c.relkind
           WHEN 'r' THEN 'table'
           WHEN 'f' THEN 'foreign'
           WHEN 'v' THEN 'view'
           WHEN 's'
               THEN 'special' END                                   AS type,
       (pg_relation_is_updatable(c.oid::regclass, FALSE) & 4) = 4   AS can_update,
       (pg_relation_is_updatable(c.oid::regclass, FALSE) & 8) = 8   AS can_insert,
       (pg_relation_is_updatable(c.oid::regclass, FALSE) & 16) = 16 AS can_delete,
       obj_description(c.oid)                                       AS comment
  FROM pg_class c
       LEFT JOIN pg_namespace n ON n.oid = c.relnamespace
 WHERE c.relkind IN ('r', 'p', 'v', 'f')
   AND n.nspname <> 'pg_catalog'
   AND n.nspname <> 'information_schema'
   AND n.nspname !~ '^pg_toast'
   AND n.nspname !~ '^pg_temp_'
   AND pg_catalog.pg_table_is_visible(c.oid)
 ORDER BY schema, name;

--# GetIndexes Index
SELECT ix.indexrelid                                                                              AS id,
       n.nspname                                                                                  AS schema,
       c.relname                                                                                  AS name,
       t.relname                                                                                  AS table,
       ix.indisunique                                                                             AS is_unique,
       ix.indisprimary                                                                            AS is_primary,
       Array(SELECT attname FROM pg_attribute a WHERE a.attrelid = ix.indexrelid ORDER BY attnum) AS columns,
       obj_description(ix.indexrelid)                                                             AS comment
  FROM pg_index ix
       JOIN pg_class t ON t.oid = ix.indrelid
       JOIN pg_class c ON c.oid = ix.indexrelid
       JOIN pg_namespace AS n
            ON n.oid = c.relnamespace
 WHERE n.nspname NOT IN ('pg_toast', 'pg_temp_1', 'pg_toast_temp_1', 'pg_catalog', 'information_schema')
 ORDER BY t.relname,
          c.relname;

--# GetEnums Enum
SELECT t.oid                                                                               AS id,
       t.typname                                                                           AS name,
       n.nspname                                                                           AS schema,
       Array(SELECT enumlabel FROM pg_enum WHERE enumtypid = t.oid ORDER BY enumsortorder) AS values,
       obj_description(t.oid)                                                              AS comment
  FROM pg_type t
       JOIN pg_namespace AS n
            ON n.oid = t.typnamespace
 WHERE t.typtype = 'e';

--# GetForeignKeys ForeignKey
SELECT con.oid                                                                         AS id,
       (SELECT nspname FROM pg_namespace ns WHERE cl.relnamespace = ns.oid)            AS schema,
       conname                                                                         AS name,
       cl.relname                                                                      AS table,
       ARRAY(SELECT attname
               FROM pg_attribute a
              WHERE a.attrelid = con.conrelid
                AND a.attnum = ANY (con.conkey)
              ORDER BY attnum)                                                          AS columns,
       (SELECT nspname FROM pg_namespace ns WHERE foreign_class.relnamespace = ns.oid) AS foreign_schema,
       foreign_class.relname                                                           AS foreign_table,
       ARRAY(SELECT attname
               FROM pg_attribute a
              WHERE a.attrelid = con.confrelid
                AND a.attnum = ANY (con.confkey)
              ORDER BY attnum)                                                          AS foreign_columns,
       obj_description(con.oid)                                                        AS comment
  FROM pg_class cl
       JOIN pg_constraint con ON con.conrelid = cl.oid
       JOIN pg_class foreign_class ON
              foreign_class.oid = con.confrelid;

--# GetColumns Column
SELECT n.nspname                                          AS schema,
       c.relname                                          AS table,
       a.attname                                          AS name,
       attnum                                             AS ordinal,
       NOT a.attnotnull                                   AS nullable,
       atthasdef                                          AS has_default,
       (SELECT typname FROM pg_type WHERE oid = atttypid) AS type,
       col_description(a.attrelid, a.attnum)              AS comment
  FROM pg_attribute a
       JOIN pg_class c ON attrelid = c.oid AND relkind IN ('v', 'r')
       JOIN pg_namespace n ON c.relnamespace = n.oid AND nspname NOT IN
                                                         ('pg_toast', 'pg_temp_1', 'pg_toast_temp_1', 'pg_catalog',
                                                          'information_schema')
 WHERE a.attnum > 0
 ORDER BY schema, "table", ordinal;
