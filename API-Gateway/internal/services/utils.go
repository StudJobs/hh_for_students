package services

import (
	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// convertRoleToGRPC конвертирует строковую роль в gRPC enum
func convertRoleToGRPC(role string) (authv1.Role, error) {
	switch role {
	case "ROLE_STUDENT":
		return authv1.Role_ROLE_STUDENT, nil
	case "ROLE_DEVELOPER":
		return authv1.Role_ROLE_DEVELOPER, nil
	case "ROLE_EMPLOYER":
		return authv1.Role_ROLE_EMPLOYER, nil
	default:
		return authv1.Role_ROLE_UNSPECIFIED, status.Error(codes.InvalidArgument, "invalid role")
	}
}

// convertRoleFromGRPC конвертирует gRPC enum в строковую роль
func convertRoleFromGRPC(role authv1.Role) string {
	switch role {
	case authv1.Role_ROLE_STUDENT:
		return "ROLE_STUDENT"
	case authv1.Role_ROLE_DEVELOPER:
		return "ROLE_DEVELOPER"
	case authv1.Role_ROLE_EMPLOYER:
		return "ROLE_EMPLOYER"
	default:
		return "ROLE_UNSPECIFIED"
	}
}
