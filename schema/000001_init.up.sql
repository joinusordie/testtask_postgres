CREATE TABLE log
(
    id SERIAL PRIMARY KEY,
    name varchar(255) not null,
    operation varchar(255) not null,
    date timestamp default current_timestamp

);