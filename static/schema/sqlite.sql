--------------------------------------------------------
--  DDL for Table dois
--------------------------------------------------------

CREATE TABLE "dois" (
    "id" INTEGER PRIMARY KEY,
    "did" VARCHAR2(700) NOT NULL UNIQUE,
    "doi" VARCHAR2(700) NOT NULL UNIQUE,
    "provider" VARCHAR2(700),
    "doiurl" VARCHAR2(700),
    "description" TEXT,
    "public" BOOLEAN,
    "metadata" BOOLEAN,
    "published" INTEGER
);
