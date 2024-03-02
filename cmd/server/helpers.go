package main

func (s *server) background(fn func()) {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				s.logger.Error("error recovering from panic", "cause", err)
			}
		}()

		fn()
	}()
}
