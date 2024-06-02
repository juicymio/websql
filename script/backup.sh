#!/bin/bash

DB_USER="root"
DB_PASSWORD="root"
DB_NAME="websql"

BACKUP_DIR="../backup"
BACKUP_DATE=$(date +%Y-%m-%d)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${BACKUP_DATE}.sql"

MYSQLDUMP_CMD="mysqldump -u${DB_USER} -p${DB_PASSWORD} ${DB_NAME}"

$MYSQLDUMP_CMD > $BACKUP_FILE
tar -czvf "${BACKUP_DIR}/${DB_NAME}_${BACKUP_DATE}.tar.gz" $BACKUP_FILE
rm $BACKUP_FILE

echo "Backup completed: ${BACKUP_DIR}/${DB_NAME}_${BACKUP_DATE}.tar.gz"
