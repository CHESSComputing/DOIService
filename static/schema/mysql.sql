CREATE TABLE dois (
    id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    did VARCHAR(700) NOT NULL UNIQUE,
    doi VARCHAR(700) NOT NULL UNIQUE,
    provider VARCHAR(700),
    doiurl VARCHAR(700),
    description TEXT,
    public BOOLEAN,
    metadata BOOLEAN,
    published INTEGER
) ENGINE=InnoDB;

-- Creating an index on doi column
CREATE INDEX idx_doi ON dois(doi);
