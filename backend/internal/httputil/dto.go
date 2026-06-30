package httputil

import (
	"time"

	"github.com/cipherkeep/backend/internal/domain"
	"github.com/cipherkeep/backend/internal/service"
)

// rfc3339 formats a time as RFC 3339 in UTC.
func rfc3339(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// rfc3339Ptr formats a nullable time as a nullable RFC 3339 string.
func rfc3339Ptr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := rfc3339(*t)
	return &s
}

// ServiceTokenMetaDTO mirrors the ServiceTokenMeta type in the API contract.
type ServiceTokenMetaDTO struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	ProjectID     string  `json:"project_id"`
	EnvironmentID string  `json:"environment_id"`
	DisplayHint   string  `json:"display_hint"`
	CreatedBy     string  `json:"created_by"`
	ExpiresAt     *string `json:"expires_at"`
	LastUsedAt    *string `json:"last_used_at"`
	CreatedAt     string  `json:"created_at"`
}

// NewServiceTokenMetaDTO maps a service token to its metadata wire representation.
func NewServiceTokenMetaDTO(t domain.ServiceToken) ServiceTokenMetaDTO {
	return ServiceTokenMetaDTO{
		ID:            t.ID,
		Name:          t.Name,
		ProjectID:     t.ProjectID,
		EnvironmentID: t.EnvironmentID,
		DisplayHint:   t.DisplayHint,
		CreatedBy:     t.CreatedBy,
		ExpiresAt:     rfc3339Ptr(t.ExpiresAt),
		LastUsedAt:    rfc3339Ptr(t.LastUsedAt),
		CreatedAt:     rfc3339(t.CreatedAt),
	}
}

// UserDTO mirrors the User type in the API contract.
type UserDTO struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// NewUserDTO maps a domain user to its wire representation.
func NewUserDTO(u *domain.User) UserDTO {
	return UserDTO{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: rfc3339(u.CreatedAt),
	}
}

// ProjectDTO mirrors the Project type in the API contract.
type ProjectDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	Role        string  `json:"role"`
	CreatedAt   string  `json:"created_at"`
}

// NewProjectDTO maps a project + role to its wire representation.
func NewProjectDTO(p domain.Project, role domain.Role) ProjectDTO {
	return ProjectDTO{
		ID:          p.ID,
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Role:        string(role),
		CreatedAt:   rfc3339(p.CreatedAt),
	}
}

// MemberDTO mirrors the Member type in the API contract.
type MemberDTO struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
}

// NewMemberDTO maps a project member to its wire representation.
func NewMemberDTO(m domain.ProjectMember) MemberDTO {
	return MemberDTO{
		UserID: m.UserID,
		Email:  m.Email,
		Name:   m.Name,
		Role:   string(m.Role),
	}
}

// EnvironmentDTO mirrors the Environment type in the API contract.
type EnvironmentDTO struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"created_at"`
}

// NewEnvironmentDTO maps an environment to its wire representation.
func NewEnvironmentDTO(e domain.Environment) EnvironmentDTO {
	return EnvironmentDTO{
		ID:        e.ID,
		ProjectID: e.ProjectID,
		Name:      e.Name,
		Slug:      e.Slug,
		CreatedAt: rfc3339(e.CreatedAt),
	}
}

// SecretMetaDTO mirrors the SecretMeta type in the API contract.
type SecretMetaDTO struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	Version   int    `json:"version"`
	UpdatedAt string `json:"updated_at"`
	UpdatedBy string `json:"updated_by"`
}

// NewSecretMetaDTO maps a secret to its metadata-only wire representation.
func NewSecretMetaDTO(s domain.Secret) SecretMetaDTO {
	return SecretMetaDTO{
		ID:        s.ID,
		Key:       s.Key,
		Version:   s.Version,
		UpdatedAt: rfc3339(s.UpdatedAt),
		UpdatedBy: s.UpdatedBy,
	}
}

// SecretDTO mirrors the Secret type (metadata + decrypted value).
type SecretDTO struct {
	SecretMetaDTO
	Value string `json:"value"`
}

// NewSecretDTO maps a decrypted secret to its full wire representation.
func NewSecretDTO(d *service.DecryptedSecret) SecretDTO {
	return SecretDTO{
		SecretMetaDTO: NewSecretMetaDTO(d.Secret),
		Value:         d.Value,
	}
}

// SecretVersionDTO mirrors the SecretVersion type in the API contract.
type SecretVersionDTO struct {
	Version   int    `json:"version"`
	CreatedAt string `json:"created_at"`
	CreatedBy string `json:"created_by"`
}

// NewSecretVersionDTO maps a secret version to its wire representation.
func NewSecretVersionDTO(v domain.SecretVersion) SecretVersionDTO {
	return SecretVersionDTO{
		Version:   v.Version,
		CreatedAt: rfc3339(v.CreatedAt),
		CreatedBy: v.CreatedBy,
	}
}

// ImportResultDTO mirrors the result of a bulk secret import.
type ImportResultDTO struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
	Total   int `json:"total"`
}

// NewImportResultDTO maps an import result to its wire representation.
func NewImportResultDTO(r *service.ImportResult) ImportResultDTO {
	return ImportResultDTO{
		Created: r.Created,
		Updated: r.Updated,
		Skipped: r.Skipped,
		Total:   r.Total,
	}
}

// ExportDTO carries serialized secrets ready for download.
type ExportDTO struct {
	Format   string `json:"format"`
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// NewExportDTO maps an export result to its wire representation.
func NewExportDTO(r *service.ExportResult) ExportDTO {
	return ExportDTO{
		Format:   r.Format,
		Filename: r.Filename,
		Content:  r.Content,
	}
}

// AuditLogDTO mirrors the AuditLog type in the API contract.
type AuditLogDTO struct {
	ID         string            `json:"id"`
	UserID     *string           `json:"user_id"`
	Action     string            `json:"action"`
	Resource   string            `json:"resource"`
	ResourceID *string           `json:"resource_id"`
	Metadata   map[string]string `json:"metadata"`
	CreatedAt  string            `json:"created_at"`
}

// NewAuditLogDTO maps an audit log to its wire representation.
func NewAuditLogDTO(l domain.AuditLog) AuditLogDTO {
	meta := l.Metadata
	if meta == nil {
		meta = map[string]string{}
	}
	return AuditLogDTO{
		ID:         l.ID,
		UserID:     l.UserID,
		Action:     l.Action,
		Resource:   l.Resource,
		ResourceID: l.ResourceID,
		Metadata:   meta,
		CreatedAt:  rfc3339(l.CreatedAt),
	}
}
