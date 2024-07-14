package service

import "go.uber.org/zap"

/*
   @Author: orbit-w
   @File: service
   @2024 4月 周日 17:53
*/

type IService interface {
	Start() error
	Stop() error
}

func Wrapper(name string) *ServiceWrapper {
	return &ServiceWrapper{
		serviceName: name,
	}
}

type ServiceWrapper struct {
	serviceName string
	start       func() error
	stop        func() error
	logger      *zap.Logger
}

func (s *ServiceWrapper) WrapStart(start func() error) *ServiceWrapper {
	s.start = start
	return s
}

func (s *ServiceWrapper) WrapStop(stop func() error) *ServiceWrapper {
	s.stop = stop
	return s
}

func (s *ServiceWrapper) WrapLogger(logger *zap.Logger) *ServiceWrapper {
	s.logger = logger
	return s
}

func (s *ServiceWrapper) Start() error {
	err := s.start()
	if err != nil {
		if s.logger != nil {
			s.logger.Error("service start error...", zap.String("Name", s.serviceName), zap.Error(err))
		}
		return err
	}

	if s.logger != nil {
		s.logger.Info("service start complete...", zap.String("Name", s.serviceName))
	}

	return nil
}

func (s *ServiceWrapper) Stop() error {
	err := s.stop()
	if err != nil {
		if s.logger != nil {
			s.logger.Error("service stop error...", zap.String("Name", s.serviceName), zap.Error(err))
		}
		return err
	}
	if s.logger != nil {
		s.logger.Info("service stop complete...", zap.String("Name", s.serviceName))
	}
	return nil
}
