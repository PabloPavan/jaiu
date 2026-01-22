-- name: GetStudent :one
SELECT * FROM students WHERE id = $1 LIMIT 1;

-- name: SearchStudents :many
SELECT *
FROM students
WHERE ($1 = '' OR full_name ILIKE '%' || $1 || '%' OR phone ILIKE '%' || $1 || '%' OR cpf ILIKE '%' || $1 || '%')
  AND (COALESCE(array_length($2::student_status[], 1), 0) = 0 OR status = ANY($2::student_status[]))
ORDER BY full_name
LIMIT $3 OFFSET $4;

-- name: CountStudents :one
SELECT COUNT(*) FROM students
WHERE ($1 = '' OR full_name ILIKE '%' || $1 || '%' OR phone ILIKE '%' || $1 || '%' OR cpf ILIKE '%' || $1 || '%')
  AND (COALESCE(array_length($2::student_status[], 1), 0) = 0 OR status = ANY($2::student_status[]));

--name: GetStudentsGrid :many
SELECT id, full_name, name, status, end_date, paid_at FROM
    (SELECT s.id,
            s.full_name,
            s.status,
            pl.name,
            su.end_date,
            su.created_at,
            p.paid_at
     FROM students as s
     LEFT JOIN subscriptions su ON s.id = su.student_id
     LEFT JOIN plans pl ON su.plan_id = pl.id
     LEFT JOIN payments p ON p.subscription_id =
         (SELECT p2.subscription_id
          FROM payments p2
          WHERE p2.subscription_id = su.id
          ORDER BY p2.paid_at DESC
          LIMIT 1))
WHERE ($1 = '' OR s.full_name ILIKE '%' || $1 || '%' OR s.phone ILIKE '%' || $1 || '%' OR s.cpf ILIKE '%' || $1 || '%')
  AND (COALESCE(array_length($2::student_status[], 1), 0) = 0 OR s.status = ANY($2::student_status[]));