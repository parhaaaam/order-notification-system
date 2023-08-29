create table agent
(
    id            serial primary key,
    slug          varchar(30) not null unique,
    current_order integer
);

create table vendor
(
    id   serial primary key,
    slug varchar(30) not null unique
);

create table "order"
(
    id            serial primary key,
    slug          varchar(50) not null,
    created_at    timestamp   not null default current_timestamp,
    time_delivery timestamp   not null,
    delivered_at  timestamp,
    vendor_id     integer,
    trip_id       integer,

    foreign key (vendor_id) references Vendor (id)
);

create table trip
(
    id        serial primary key,
    order_id  integer unique,
    vendor_id integer,
    status    varchar(30) not null,

    foreign key (order_id) references "order" (id),
    foreign key (vendor_id) references Vendor (id)

);

create table delay_report
(
    id          serial primary key,
    description varchar(300),
    order_id    integer,
    agent_id    integer,
    status      varchar(30),
    created_at  timestamp not null default current_timestamp,

    foreign key (order_id) references "order" (id),
    foreign key (agent_id) references Agent (id)
)
