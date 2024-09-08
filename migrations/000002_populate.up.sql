INSERT INTO users (username, email, password_hash)
VALUES
    ('john_doe', 'john@example.com', '$2a$12$l4JS9.kv0OS.sp3kuGna7OiUuN4hjTp/wW.oMOV5ix6FJglQFevpO'),  
    ('jane_smith', 'jane@example.com', '$2a$12$l4JS9.kv0OS.sp3kuGna7OiUuN4hjTp/wW.oMOV5ix6FJglQFevpO'),
    ('alice_wonder', 'alice@example.com', '$2a$12$l4JS9.kv0OS.sp3kuGna7OiUuN4hjTp/wW.oMOV5ix6FJglQFevpO');  

INSERT INTO posts (user_id, title, content)
VALUES
    (1, 'Welcome to My Blog', 'This is the first post on my new blog!'),
    (2, 'A Day in the Life', 'Today I went to the park and had a great time.'),
    (3, 'Tech Trends 2024', 'Lets talk about the top tech trends for 2024.');

INSERT INTO comments (post_id, user_id, content)
VALUES
    (1, 2, 'Great first post! Looking forward to more.'),
    (1, 3, 'Welcome to the blogging world!'),
    (2, 1, 'Sounds like a fun day!'),
    (3, 2, 'I’m excited about AI developments.'),
    (3, 1, 'This is very insightful, thanks for sharing.');

INSERT INTO tags (name)
VALUES ('Technology'),
       ('Programming'),
       ('Science'),
       ('Health'),
       ('Travel'),
       ('Education');

INSERT INTO post_tags (post_id, tag_id)
VALUES (1, 1), 
       (1, 2),
       (2, 3),
       (3, 4),
       (3, 5);
