ALTER TABLE contracts
ADD COLUMN customer_address VARCHAR(100) not null default '',
ADD COLUMN performer_address VARCHAR(100) not null default '',
ADD COLUMN contract_address VARCHAR(100) not null default '';

comment on column contracts.customer_address is 'Customer address in in block chain';
comment on column contracts.performer_address is 'Performer address in in block chain';
comment on column contracts.contract_address is 'Address in the block chain relevant smart contract';
