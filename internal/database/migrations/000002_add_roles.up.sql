INSERT INTO "roles"
("display_name", "description")
VALUES (
    'staff',
    'An employee from HR with permissions to solve'
    || 'tickets and check attendance records of all employees'
),
(
    'leader',
    'An employee on charge of a team with permissions to'
    || 'solve tickets and check attendance records of their'
    || 'team members'
),
(
    'employee',
    'The base role of an employee with permissions to open'
    || 'tickets and check their own attendance records'
);
