INSERT INTO "roles" ("code")
VALUES ('admin'),
('staff'),
('leader'),
('employee');

INSERT INTO "permissions" ("code", "description")
VALUES (
    'attendances:create:own',
    'permite a criação de registros de ponto para o próprio usuário'
),
(
    'attendances:read:own',
    'permite a consulta de registros de ponto do próprio usuário'
),
(
    'attendances:read:any',
    'permite a consulta de registros de ponto de outros usuários'
),
(
    'tickets:write:own',
    'permite a abertura de tickets para ajuste de registros de ponto'
    || 'ou fechamento de tickets abertos pelo próprio usuário'
),
(
    'tickets:read:any',
    'permite a consulta de tickets abertos por outros usuários'
),
(
    'tickets:update:any',
    'permite a resolução de tickets abertos por outros usuários'
),
(
    'users:write:any',
    'permite o cadastro de novos usuários no sistema ou atualização de'
    || 'informações pessoais de outros usuários como `name` and `status`'
),
(
    'users:read:any',
    'permite a leitura de informações do próprio usuário e'
    || 'de outros usuários na plataforma'
),
(
    'tokens:write:own',
    'permite a criação de tokens para diferentes'
    || 'escopos (ativação ou autenticação)'
),
(
    'credentials:write:own',
    'permite ao usuário a criação ou atualização de credenciais para login'
),
(
    'profiles:write:own',
    'permite a criação  ou edição de informações de perfil do próprio usuário'
),
(
    'profiles:read:any',
    'permite a consulta de informações de perfil do próprio usuário ou de'
    || 'outros usuários da plataforma'
),
(
    'users_roles:write:any',
    'permite a atribuição ou deleção de roles de outros usuários. Um usuário'
    || 'nunca pode editar a própria role e a role `employee` não pode ser'
    || 'deletada de nenhum usuário'
),
(
    'users_roles:read:own',
    'permite a consulta das roles do próprio usuário'
),
(
    'users_roles:read:any',
    'permite a consulta das roles de outros usuários'
),
(
    'permissions:read:own',
    'permite a consulta das permissões do próprio usuário'
),
(
    'permissions:read:any',
    'permite a consulta das permissões  de outros usuários'
);

INSERT INTO "permissions_roles" ("role_id", "permission_id")
VALUES (
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'admin'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'users_roles:write:any'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'admin'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'permissions:read:any'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'staff'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'attendances:read:any'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'staff'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'tickets:read:any'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'staff'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'tickets:update:any'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'leader'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'attendances:read:any'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'leader'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'tickets:read:any'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'leader'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'tickets:update:any'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'employee'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'attendances:create:own'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'employee'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'attendances:read:own'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'employee'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'tickets:write:own'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'employee'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'users:read:any'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'employee'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'tokens:write:own'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'employee'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'credentials:write:own'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'employee'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'profiles:write:own'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'employee'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'profiles:read:any'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'employee'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'users_roles:read:own'
    )
),
(
    (
        SELECT "id"
        FROM "roles"
        WHERE "code" = 'employee'
    ),
    (
        SELECT "id"
        FROM "permissions"
        WHERE "code" = 'permissions:read:own'
    )
);
