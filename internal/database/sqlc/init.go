package database

// createDefaultAdmin creates the default admin user if it doesn't exist.
func createDefaultAdmin() error {
	// 	var cfg config
	//
	// 	flag.StringVar(&cfg.db.dsn, "db-dsn", "user:pass@localhost:5432/db", "postgreSQL DSN")
	//
	// 	flag.StringVar(&cfg.rootUser.name, "root-name", "Root", "The name of the root user")
	// 	flag.StringVar(&cfg.rootUser.email, "root-email", "root@example.com", "The email of the root user")
	// 	flag.StringVar(&cfg.rootUser.password, "root-password", "root", "The password of the root user")
	//
	// 	flag.Func("roles", "The roles to be created, separated by spaces", func(val string) error {
	// 		cfg.roles = strings.Fields(val)
	// 		return nil
	// 	})
	//
	// 	flag.Parse()
	//
	// 	store, err := database.NewStore(cfg.db.dsn)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	//
	// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 	defer cancel()
	//
	// 	err = store.ExecTx(ctx, func(q *database.Queries) error {
	// 		hashedPassword, err := password.Hash(cfg.rootUser.password)
	// 		if err != nil {
	// 			return err
	// 		}
	//
	// 		employee, err := q.CreateEmployee(ctx, database.CreateEmployeeParams{
	// 			Name:  cfg.rootUser.name,
	// 			Email: cfg.rootUser.email,
	// 			HashedPassword: pgtype.Text{
	// 				String: hashedPassword,
	// 				Valid:  true,
	// 			},
	// 			Status: "active",
	// 		})
	// 		if err != nil {
	// 			return err
	// 		}
	//
	// 		_, err = q.CreateRoles(ctx, cfg.roles)
	// 		if err != nil {
	// 			return err
	// 		}
	//
	// 		roles, err := q.GetRoles(ctx)
	// 		if err != nil {
	// 			return err
	// 		}
	//
	// 		var args []database.AddRolesForEmployeeParams
	//
	// 		for _, r := range roles {
	// 			args = append(args, database.AddRolesForEmployeeParams{
	// 				EmployeeID: employee.ID,
	// 				RoleID:     r.ID,
	// 				Grantor:    employee.ID,
	// 			})
	// 		}
	//
	// 		_, err = q.AddRolesForEmployee(ctx, args)
	// 		if err != nil {
	// 			return err
	// 		}
	//
	// 		return nil
	// 	})
	//
	// 	return err
	// }
	//
	// func main() {
	// 	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	//
	// 	err := run()
	// 	if err != nil {
	// 		trace := string(debug.Stack())
	// 		logger.Error(err.Error(), "trace", trace)
	// 		os.Exit(1)
	// 	}
	//
	// 	logger.Debug("database initialized")
	return nil
}
