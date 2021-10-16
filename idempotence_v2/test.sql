CREATE DATABASE test;

CREATE TABLE `student` (
    `id` int NOT NULL AUTO_INCREMENT,
    `idempotence_id` varchar(60) DEFAULT '',
    `name` varchar(10) DEFAULT '',
    `age` int DEFAULT 0,
    `money` int DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idempotence_id`
) ENGINE=InnoDB
