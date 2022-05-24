alter table persons
add column ethereum_address varchar(100) not null default '';

comment on column persons.ethereum_address is 'Person address in Ethereum-compatible block chains';