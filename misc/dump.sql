CREATE
    DATABASE IF NOT EXISTS xxxxxxxx DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;

USE xxxxxxxx;

CREATE TABLE `back_transaction`
(
    `file_name` varchar(63) NOT NULL,
    `dir_path` varchar(127) DEFAULT NULL,
    `status` varchar(20) NOT NULL,
    `timestamp` timestamp,

    PRIMARY KEY (`file_name`)
);