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
  photo_url,
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
  photo_url = $10,
  status = $11,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetStudent :one
SELECT * FROM students WHERE id = $1 LIMIT 1;

-- name: SearchStudents :many
SELECT *
FROM students
WHERE ($1 = '' OR full_name ILIKE '%' || $1 || '%' OR phone ILIKE '%' || $1 || '%' OR cpf ILIKE '%' || $1 || '%')
  AND (COALESCE(array_length($2::text[], 1), 0) = 0 OR status = ANY($2::text[]))
ORDER BY full_name
LIMIT $3 OFFSET $4;
