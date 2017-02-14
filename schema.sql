DROP TABLE IF EXISTS files;
CREATE TABLE files (
    name TEXT PRIMARY KEY,
    email TEXT, -- email of uploader
    token TEXT, -- token sent via email
    authorized INTEGER, -- either 0 or 1
    uploaded INTEGER -- datetime the file was uploaded
);