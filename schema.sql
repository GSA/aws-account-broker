CREATE TABLE "service_instances" ("id" integer primary key autoincrement,"created_at" datetime,"updated_at" datetime,"deleted_at" datetime,"instance_id" varchar(255),"request_id" varchar(255) );
CREATE INDEX idx_service_instances_deleted_at ON "service_instances"(deleted_at) ;
