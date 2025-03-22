--------------------------------------------------------
--  DDL for Table dois
--------------------------------------------------------

CREATE TABLE "dois" (
    "id" INTEGER PRIMARY KEY,
    "did" VARCHAR2(700) NOT NULL UNIQUE,
    "doi" VARCHAR2(700) NOT NULL UNIQUE,
    "description" TEXT,
    "metadata" TEXT,
    "published" INTEGER
);
