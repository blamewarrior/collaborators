package blamewarrior

const (
	CreateRepositoryQuery = `
    INSERT INTO repositories(full_name) VALUES($1) RETURNING id
  `
)
