CREATE DATABASE acl;

USE acl;

CREATE TABLE IF NOT EXISTS `services` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(32) NOT NULL UNIQUE,
  `description` varchar(255) NOT NULL,
  `status` varchar(16) NOT NULL DEFAULT 'enable',
  `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 AUTO_INCREMENT=1;

CREATE TABLE IF NOT EXISTS `items` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `service_id` int(11) NOT NULL,
  `name` varchar(32) NOT NULL,
  `description` varchar(255) NOT NULL,
  `protocol` varchar(16) NOT NULL,
  `cidr` varchar(32) NOT NULL,
  `port` varchar(16) NOT NULL,
  `status` varchar(16) NOT NULL DEFAULT 'enable',
  `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY (`id`),
   FOREIGN KEY (service_id)  REFERENCES services (id)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 AUTO_INCREMENT=11;

CREATE TABLE IF NOT EXISTS `rules` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `source` int(11) NOT NULL,
  `destination` int(11) NOT NULL,
  `name` varchar(32) NOT NULL,
  `description` varchar(255) NOT NULL,
  `status` varchar(16) NOT NULL DEFAULT 'enable',
  `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY (`id`),
   FOREIGN KEY (source)  REFERENCES services (id),
   FOREIGN KEY (destination)  REFERENCES services (id)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 AUTO_INCREMENT=1;

CREATE TABLE IF NOT EXISTS `events` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `type` varchar(32) NOT NULL,
  `event` varchar(255) NOT NULL,
  `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 AUTO_INCREMENT=1;




CREATE USER acluser@localhost IDENTIFIED BY 'aclpass';
CREATE USER 'acluser'@'localhost' IDENTIFIED WITH mysql_native_password BY 'aclpass';
GRANT ALL PRIVILEGES ON acl.* TO 'acluser'@'localhost';
FLUSH PRIVILEGES;

insert into services (name, description, status)  values ('test srv1', 'test service1', 'enabled');
insert into services (name, description, status)  values ('test srv2', 'test service2', 'enabled');

insert into items (service_id, name, description, protocol, cidr, port, status)  values ('1', 'items1', 'test items 1', 'tcp', '10.10.1.1', '441', 'enabled');
insert into items (service_id, name, description, protocol, cidr, port, status)  values ('2', 'items2', 'test items 2', 'tcp', '10.10.2.1', '442', 'enabled');

insert into rules (source, destination, name, description, status)  values ('1', '2', 'rule1-2', 'test rule 1-2', 'enabled');




SELECT i.id, s.id AS sid, s.name AS sname, i.name, i.description, i.protocol, i.cidr, i.port, i.status Sname 
FROM items i JOIN services s ON i.service_id=s.id ORDER BY id;

SELECT r.id, s.id AS sid, s.name AS sname, d.id AS did, d.name AS dname, r.name, r.description, r.status, r.created, r.updated 
FROM rules r JOIN services s ON r.source=s.id JOIN services d ON r.destination=d.id ORDER BY id;

SELECT s.id AS sid, s.name AS sname, d.id AS did, d.name AS dname FROM rules r JOIN services s ON r.source=s.id JOIN services d ON r.destination=d.id;


SELECT d.protocol, s.cidr AS source, d.cidr AS destination, d.port, r.description 
FROM rules r JOIN items s ON r.source=s.service_id JOIN items d ON r.destination=d.service_id;

SELECT d.protocol, s.cidr AS source, d.cidr AS destination, d.port, r.name, s.name AS sitem, d.name AS ditem,  sr.name AS sservice, dr.name AS dservice 
FROM rules r JOIN items s ON r.source=s.service_id JOIN items d ON r.destination=d.service_id JOIN services sr ON r.source=sr.id JOIN services dr ON r.destination=dr.id;

SELECT d.protocol, s.cidr AS source, d.cidr AS destination, d.port, r.name, s.name AS sitem, d.name AS ditem,  sr.name AS sservice, dr.name AS dservice  
FROM rules r JOIN items s ON r.source=s.service_id JOIN items d ON r.destination=d.service_id JOIN services sr ON r.source=sr.id JOIN services dr ON r.destination=dr.id 
WHERE s.status='enabled' AND d.status='enabled' AND sr.status='enabled' AND dr.status='enabled' AND r.status='enabled';

SELECT d.protocol, s.cidr AS source, d.cidr AS destination, d.port, 
r.id AS rid, r.name AS rname, s.id AS siid, s.name AS sitem, d.id AS diid, d.name AS ditem, sr.id AS ssid, sr.name AS sservice, dr.id AS dsid, dr.name AS dservice 
FROM rules r JOIN items s ON r.source=s.service_id JOIN items d ON r.destination=d.service_id JOIN services sr ON r.source=sr.id JOIN services dr ON r.destination=dr.id 
WHERE d.protocol LIKE '%10.10.1.1%' OR s.cidr LIKE '%10.10.1.1%' OR d.cidr LIKE '%10.10.1.1%' OR d.port LIKE '%10.10.1.1%' OR r.name LIKE '%10.10.1.1%' 
OR s.name LIKE '%10.10.1.1%' OR d.name LIKE '%10.10.1.1%' OR sr.name LIKE '%10.10.1.1%' OR dr.name LIKE '%10.10.1.1%';


SELECT * FROM services WHERE name LIKE '%enabled%' OR description LIKE '%enabled%' OR status LIKE '%enabled%';
