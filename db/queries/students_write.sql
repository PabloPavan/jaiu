-- name: CreateStudent :one
INSERT INTO students (
  full_name,
  birth_date,
  gender,
  phone,
  email,
  cpf,
  address,
  notes,
  photo_object_key,
  status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdateStudent :one
UPDATE students
SET
  full_name = $2,
  birth_date = $3,
  gender = $4,
  phone = $5,
  email = $6,
  cpf = $7,
  address = $8,
  notes = $9,
  photo_object_key = $10,
  status = $11,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetStudent :one
SELECT * FROM students WHERE id = $1 LIMIT 1;