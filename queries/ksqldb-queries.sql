-- create messages_stream
CREATE STREAM messages_stream (
       user_id      BIGINT,
       recipient_id BIGINT,
       message      STRING,
       created_at   STRING
) WITH (
       KAFKA_TOPIC='messages',
       VALUE_FORMAT='JSON_SR',
       PARTITIONS=3
);

-- show messages
SELECT *
  FROM messages_stream
  EMIT CHANGES
 LIMIT 5;

-- create table with count
CREATE TABLE total_messages_count AS
SELECT 'all' AS dummy_key,
       COUNT(*) AS message_count
  FROM messages_stream
 GROUP BY 'all'
  EMIT CHANGES;

-- show message count
SELECT message_count
  FROM total_messages_count
  EMIT CHANGES
 LIMIT 1;

-- create table with unique recipient
CREATE TABLE unique_recipients_count AS
SELECT 'all' AS dummy_key,
       COUNT_DISTINCT(recipient_id) AS recipient_count
  FROM messages_stream
 GROUP BY 'all'
  EMIT CHANGES;

-- show recipient count
SELECT recipient_count
  FROM unique_recipients_count
  EMIT CHANGES
 LIMIT 1;

-- create table with top message sender
CREATE TABLE messages_by_user AS
SELECT user_id,
       COUNT(*) AS message_count
  FROM messages_stream
 GROUP BY user_id
  EMIT CHANGES;

-- show 5 top message sender (but emitter use only 3 sender)
SELECT user_id,
       message_count
  FROM messages_by_user
  EMIT CHANGES
 LIMIT 5;

-- create table with user statistics
CREATE TABLE user_statistics AS
SELECT user_id,
       COUNT(*) AS message_count,
       COUNT_DISTINCT(recipient_id) AS recipient_count
  FROM messages_stream
 GROUP BY user_id
  EMIT CHANGES;

-- show user statistics
SELECT *
  FROM user_statistics
PARTITION BY user_id
  EMIT CHANGES;