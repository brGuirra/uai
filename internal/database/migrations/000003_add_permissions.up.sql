INSERT INTO "permissions" ("display_name", "description") VALUES
(
    'admin',
    'Allows full access to the system, including the ability'
    || 'to promote other users to admin level'
),
('user_manager', 'Allows an user to create and edit another users'),
(
    'ticket_manager', 'Allows an user to view, solve and close tickets'
),
(
    'ticket_issuer', 'Allows an user to open, view and close their tickets'
    || 'own tickets'
),
('attendance_manager', 'Allows an user to view and edit attendance records'),
(
    'attendance_registration',
    'Allows an user to register and view their own attendance'
    || 'records'
);
