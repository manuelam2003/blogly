-- Delete the comments inserted in the 'populate.sql' script
DELETE FROM comments
WHERE content IN (
    'Great first post! Looking forward to more.',
    'Welcome to the blogging world!',
    'Sounds like a fun day!',
    'I’m excited about AI developments.',
    'This is very insightful, thanks for sharing.'
);

-- Delete the posts inserted in the 'populate.sql' script
DELETE FROM posts
WHERE title IN (
    'Welcome to My Blog',
    'A Day in the Life',
    'Tech Trends 2024'
);

-- Delete the users inserted in the 'populate.sql' script
DELETE FROM users
WHERE username IN (
    'john_doe',
    'jane_smith',
    'alice_wonder'
);
