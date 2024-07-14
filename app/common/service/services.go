package service

/*
   @Author: orbit-w
   @File: services
   @2024 4月 周日 18:25
*/

type Services struct {
	services []IService
}

func NewServices() *Services {
	return &Services{
		services: make([]IService, 0),
	}
}

func (s *Services) Reg(service IService) *Services {
	s.services = append(s.services, service)
	return s
}

func (s *Services) Start() error {
	var (
		err     error
		started = make([]IService, 0)
	)

	defer func() {
		if err != nil {
			for i := range started {
				_ = started[i].Stop()
			}
		}
	}()

	for i := range s.services {
		service := s.services[i]
		err = service.Start()
		if err != nil {
			return err
		}
		started = append(started, service)
	}
	return nil
}

func (s *Services) Stop() {
	length := len(s.services)
	if length > 0 {
		for i := length - 1; i >= 0; i-- {
			service := s.services[i]
			_ = service.Stop()
		}
	}
}
