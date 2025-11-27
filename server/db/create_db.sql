drop table if exists user;
drop table if exists oauth_account;

create table user (
    id bigint AUTO_INCREMENT PRIMARY KEY,
    email varchar(255) NOT NULL UNIQUE,
    password_hash varchar(255) DEFAULT NULL,
    full_name varchar(255) DEFAULT NULL,
    phone varchar(255) DEFAULT NULL,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp on update current_timestamp
);

create table oauth_account (
    id bigint AUTO_INCREMENT PRIMARY KEY,
    user_id bigint NOT NULL,
    provider_name varchar(255) NOT NULL,
    sub varchar(255) NOT NULL,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp on update current_timestamp,
    foreign key (user_id) references user(id) on delete cascade,
    unique key (provider_name, sub)
);
