CREATE TABLE todos
(
    ID SERIAL PRIMARY KEY,
    Title VARCHAR(50) NOT NULL,
    Description TEXT,
    Tags VARCHAR(50)[],
    DueTime date
)