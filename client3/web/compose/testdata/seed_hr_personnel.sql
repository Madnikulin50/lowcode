-- HR personnel dashboard: employees, leave, vacancies + hr_slice for charts/KPI.
-- Self-contained synthetic data (no dependency on other seed tables).

DROP TABLE IF EXISTS hr_slice CASCADE;
DROP TABLE IF EXISTS hr_leave CASCADE;
DROP TABLE IF EXISTS hr_vacancy CASCADE;
DROP TABLE IF EXISTS hr_employee CASCADE;

CREATE TABLE hr_employee (
  id              bigserial PRIMARY KEY,
  emp_code        text NOT NULL,
  full_name       text NOT NULL,
  department      text NOT NULL,
  title           text NOT NULL,
  status          text NOT NULL,  -- active | probation | on_leave | terminated
  employment_type text NOT NULL,  -- full_time | part_time | contract
  hire_date       date NOT NULL,
  tenure_years    numeric(6,1) NOT NULL DEFAULT 0,
  salary_band     text NOT NULL,  -- A | B | C | D
  manager_name    text,
  location        text NOT NULL,
  email           text
);

CREATE TABLE hr_leave (
  id              bigserial PRIMARY KEY,
  emp_code        text NOT NULL,
  full_name       text NOT NULL,
  department      text NOT NULL,
  leave_type      text NOT NULL,  -- vacation | sick | unpaid | parental
  start_date      date NOT NULL,
  end_date        date NOT NULL,
  days            numeric(6,1) NOT NULL,
  status          text NOT NULL     -- approved | pending | rejected
);

CREATE TABLE hr_vacancy (
  id              bigserial PRIMARY KEY,
  position_title  text NOT NULL,
  department      text NOT NULL,
  location        text NOT NULL,
  employment_type text NOT NULL,
  openings        numeric(6,0) NOT NULL DEFAULT 1,
  priority        text NOT NULL,   -- high | medium | low
  days_open       numeric(6,0) NOT NULL DEFAULT 0,
  hiring_manager  text
);

-- Departments / roles used to generate a realistic headcount mix
WITH depts(department, titles) AS (
  VALUES
    ('Розница', ARRAY['Продавец','Старший продавец','Администратор магазина','Кассир']),
    ('Склад', ARRAY['Кладовщик','Комплектовщик','Начальник смены','Водитель']),
    ('Офис', ARRAY['Бухгалтер','Экономист','Офис-менеджер','Юрист']),
    ('IT', ARRAY['Разработчик','Аналитик','DevOps','Руководитель IT']),
    ('Маркетинг', ARRAY['Маркетолог','Контент-менеджер','Дизайнер','SMM']),
    ('HR', ARRAY['HR-менеджер','Рекрутер','Специалист по кадрам']),
    ('Логистика', ARRAY['Логист','Диспетчер','Координатор поставок']),
    ('Безопасность', ARRAY['Охранник','Специалист СБ','Руководитель СБ'])
),
locs(location) AS (
  VALUES ('Москва'), ('Санкт-Петербург'), ('Казань'), ('Новосибирск'), ('Екатеринбург')
),
names(n) AS (
  VALUES
    ('Иванов Алексей'), ('Петрова Мария'), ('Сидоров Дмитрий'), ('Козлова Анна'),
    ('Смирнов Иван'), ('Васильева Елена'), ('Морозов Павел'), ('Новикова Ольга'),
    ('Фёдоров Николай'), ('Волкова Дарья'), ('Алексеев Сергей'), ('Лебедева Ирина'),
    ('Семёнов Андрей'), ('Егорова Наталья'), ('Павлов Максим'), ('Кузнецова Юлия'),
    ('Соколов Артём'), ('Михайлова Виктория'), ('Зайцев Роман'), ('Белова Екатерина'),
    ('Орлов Кирилл'), ('Андреева София'), ('Макаров Егор'), ('Никитина Алина'),
    ('Захаров Тимур'), ('Борисова Полина'), ('Тарасов Владислав'), ('Громова Ксения'),
    ('Куликов Денис'), ('Савельева Маргарита'), ('Баранов Илья'), ('Тихонова Вера'),
    ('Комаров Георгий'), ('Беляева Людмила'), ('Щербаков Степан'), ('Данилова Татьяна'),
    ('Медведев Олег'), ('Жукова Анастасия'), ('Крылов Вадим'), ('Ларина Марина'),
    ('Соловьёв Глеб'), ('Фролова Яна'), ('Гусев Борис'), ('Матвеева Диана'),
    ('Титов Руслан'), ('Романова Валерия'), ('Белоусов Артём'), ('Карпова Елизавета'),
    ('Субботин Пётр'), ('Власова Арина'), ('Ершов Константин'), ('Сазонова Ульяна'),
    ('Гордеев Марк'), ('Панфилова Ника'), ('Лазарев Савелий'), ('Чернова Агата')
),
gen AS (
  SELECT
    row_number() OVER () AS rn,
    n.n AS full_name,
    d.department,
    d.titles[1 + ((row_number() OVER ()) % cardinality(d.titles))] AS title,
    l.location
  FROM names n
  CROSS JOIN LATERAL (
    SELECT * FROM depts
    ORDER BY md5(n.n || department)
    LIMIT 1
  ) d
  CROSS JOIN LATERAL (
    SELECT * FROM locs
    ORDER BY md5(n.n || location)
    LIMIT 1
  ) l
)
INSERT INTO hr_employee (
  emp_code, full_name, department, title, status, employment_type,
  hire_date, tenure_years, salary_band, manager_name, location, email
)
SELECT
  'E' || lpad(rn::text, 4, '0'),
  full_name,
  department,
  title,
  CASE
    WHEN rn % 17 = 0 THEN 'terminated'
    WHEN rn % 11 = 0 THEN 'on_leave'
    WHEN rn % 9 = 0 THEN 'probation'
    ELSE 'active'
  END,
  CASE
    WHEN rn % 13 = 0 THEN 'contract'
    WHEN rn % 7 = 0 THEN 'part_time'
    ELSE 'full_time'
  END,
  CURRENT_DATE - (((rn * 37) % 2800 + 30)::integer),
  ROUND((((rn * 37) % 2800 + 30) / 365.0)::numeric, 1),
  CASE
    WHEN rn % 10 = 0 THEN 'D'
    WHEN rn % 5 = 0 THEN 'C'
    WHEN rn % 3 = 0 THEN 'B'
    ELSE 'A'
  END,
  CASE
    WHEN department IN ('Розница', 'Склад') THEN 'Региональный директор'
    WHEN department = 'IT' THEN 'CTO'
    ELSE 'Руководитель отдела'
  END,
  location,
  lower(regexp_replace(split_part(full_name, ' ', 1), '[^a-zA-Zа-яА-ЯёЁ]', '', 'g'))
    || rn::text || '@example.local'
FROM gen;

INSERT INTO hr_leave (emp_code, full_name, department, leave_type, start_date, end_date, days, status)
SELECT
  e.emp_code,
  e.full_name,
  e.department,
  (ARRAY['vacation','sick','unpaid','parental'])[1 + (e.id % 4)],
  CASE
    WHEN e.status = 'on_leave' THEN CURRENT_DATE - (((e.id % 5) + 1)::integer)
    ELSE CURRENT_DATE + (((e.id % 20) + 3)::integer)
  END,
  CASE
    WHEN e.status = 'on_leave' THEN CURRENT_DATE + (((e.id % 7) + 2)::integer)
    ELSE CURRENT_DATE + (((e.id % 20) + 10)::integer)
  END,
  CASE
    WHEN e.id % 4 = 0 THEN 14
    WHEN e.id % 4 = 1 THEN 3
    WHEN e.id % 4 = 2 THEN 5
    ELSE 28
  END,
  CASE
    WHEN e.status = 'on_leave' THEN 'approved'
    WHEN e.id % 5 = 0 THEN 'pending'
    ELSE 'approved'
  END
FROM hr_employee e
WHERE e.status IN ('active', 'probation', 'on_leave')
  AND (e.status = 'on_leave' OR e.id % 4 = 0);

INSERT INTO hr_vacancy (position_title, department, location, employment_type, openings, priority, days_open, hiring_manager)
VALUES
  ('Продавец', 'Розница', 'Москва', 'full_time', 3, 'high', 21, 'Региональный директор'),
  ('Администратор магазина', 'Розница', 'Казань', 'full_time', 1, 'high', 35, 'Региональный директор'),
  ('Кладовщик', 'Склад', 'Москва', 'full_time', 2, 'medium', 14, 'Начальник склада'),
  ('Разработчик', 'IT', 'Санкт-Петербург', 'full_time', 2, 'high', 45, 'CTO'),
  ('Аналитик', 'IT', 'Москва', 'full_time', 1, 'medium', 28, 'CTO'),
  ('Маркетолог', 'Маркетинг', 'Москва', 'full_time', 1, 'low', 10, 'CMO'),
  ('Рекрутер', 'HR', 'Москва', 'full_time', 1, 'medium', 18, 'HRD'),
  ('Логист', 'Логистика', 'Новосибирск', 'full_time', 1, 'high', 40, 'Директор логистики'),
  ('Бухгалтер', 'Офис', 'Екатеринбург', 'full_time', 1, 'medium', 12, 'Финдиректор'),
  ('Охранник', 'Безопасность', 'Москва', 'shift', 2, 'low', 7, 'Руководитель СБ'),
  ('Кассир', 'Розница', 'Санкт-Петербург', 'part_time', 2, 'medium', 9, 'Региональный директор'),
  ('DevOps', 'IT', 'Москва', 'contract', 1, 'high', 52, 'CTO');

CREATE TABLE hr_slice AS
WITH active AS (
  SELECT * FROM hr_employee WHERE status IN ('active', 'probation', 'on_leave')
),
u AS (
  -- KPI row
  SELECT
    'kpi'::text AS slice_kind,
    'all'::text AS slice_key,
    'KPI'::text AS slice_label,
    (SELECT COUNT(*)::numeric FROM active) AS headcount,
    (SELECT COUNT(*)::numeric FROM hr_employee WHERE status = 'active') AS active_count,
    (SELECT COUNT(*)::numeric FROM hr_employee WHERE status = 'probation') AS probation_count,
    (SELECT COUNT(*)::numeric FROM hr_employee WHERE status = 'on_leave') AS on_leave_count,
    (SELECT COUNT(*)::numeric FROM hr_employee WHERE status = 'terminated') AS terminated_count,
    (SELECT COALESCE(SUM(openings), 0)::numeric FROM hr_vacancy) AS open_positions,
    (SELECT COUNT(*)::numeric FROM hr_leave WHERE status = 'pending') AS pending_leave,
    (SELECT COUNT(*)::numeric FROM hr_leave
      WHERE status = 'approved' AND start_date <= CURRENT_DATE AND end_date >= CURRENT_DATE) AS currently_away,
    (SELECT ROUND(
        100.0 * COUNT(*) FILTER (WHERE status = 'terminated')
        / NULLIF(COUNT(*), 0), 1
      ) FROM hr_employee WHERE hire_date >= CURRENT_DATE - INTERVAL '365 days'
        OR status = 'terminated') AS turnover_pct,
    (SELECT ROUND(AVG(tenure_years), 1) FROM active) AS avg_tenure,
    0::numeric AS openings,
    NULL::text AS extra_label

  UNION ALL

  SELECT
    'department',
    department,
    department,
    COUNT(*)::numeric,
    COUNT(*) FILTER (WHERE status = 'active')::numeric,
    COUNT(*) FILTER (WHERE status = 'probation')::numeric,
    COUNT(*) FILTER (WHERE status = 'on_leave')::numeric,
    COUNT(*) FILTER (WHERE status = 'terminated')::numeric,
    0, 0, 0, 0,
    ROUND(AVG(tenure_years) FILTER (WHERE status IN ('active','probation','on_leave')), 1),
    0,
    NULL
  FROM hr_employee
  GROUP BY department

  UNION ALL

  SELECT
    'status',
    status,
    CASE status
      WHEN 'active' THEN 'Работает'
      WHEN 'probation' THEN 'Испытательный'
      WHEN 'on_leave' THEN 'В отпуске'
      WHEN 'terminated' THEN 'Уволен'
      ELSE status
    END,
    COUNT(*)::numeric,
    COUNT(*) FILTER (WHERE status = 'active')::numeric,
    COUNT(*) FILTER (WHERE status = 'probation')::numeric,
    COUNT(*) FILTER (WHERE status = 'on_leave')::numeric,
    COUNT(*) FILTER (WHERE status = 'terminated')::numeric,
    0, 0, 0, 0, 0, 0, NULL
  FROM hr_employee
  GROUP BY status

  UNION ALL

  SELECT
    'employment',
    employment_type,
    CASE employment_type
      WHEN 'full_time' THEN 'Полная занятость'
      WHEN 'part_time' THEN 'Частичная'
      WHEN 'contract' THEN 'Договор'
      WHEN 'shift' THEN 'Сменный'
      ELSE employment_type
    END,
    COUNT(*)::numeric,
    COUNT(*) FILTER (WHERE status = 'active')::numeric,
    COUNT(*) FILTER (WHERE status = 'probation')::numeric,
    COUNT(*) FILTER (WHERE status = 'on_leave')::numeric,
    COUNT(*) FILTER (WHERE status = 'terminated')::numeric,
    0, 0, 0, 0, 0, 0, NULL
  FROM hr_employee
  WHERE status IN ('active', 'probation', 'on_leave')
  GROUP BY employment_type

  UNION ALL

  SELECT
    'location',
    location,
    location,
    COUNT(*)::numeric,
    COUNT(*) FILTER (WHERE status = 'active')::numeric,
    COUNT(*) FILTER (WHERE status = 'probation')::numeric,
    COUNT(*) FILTER (WHERE status = 'on_leave')::numeric,
    COUNT(*) FILTER (WHERE status = 'terminated')::numeric,
    0, 0, 0, 0,
    ROUND(AVG(tenure_years) FILTER (WHERE status IN ('active','probation','on_leave')), 1),
    0,
    NULL
  FROM hr_employee
  WHERE status IN ('active', 'probation', 'on_leave')
  GROUP BY location

  UNION ALL

  SELECT
    'tenure',
    band,
    label,
    cnt,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, NULL
  FROM (
    SELECT
      CASE
        WHEN tenure_years < 1 THEN '0-1'
        WHEN tenure_years < 3 THEN '1-3'
        WHEN tenure_years < 5 THEN '3-5'
        ELSE '5+'
      END AS band,
      CASE
        WHEN tenure_years < 1 THEN 'до 1 года'
        WHEN tenure_years < 3 THEN '1–3 года'
        WHEN tenure_years < 5 THEN '3–5 лет'
        ELSE '5+ лет'
      END AS label,
      COUNT(*)::numeric AS cnt
    FROM hr_employee
    WHERE status IN ('active', 'probation', 'on_leave')
    GROUP BY 1, 2
  ) t

  UNION ALL

  SELECT
    'vacancy_dept',
    department,
    department,
    SUM(openings)::numeric,
    0, 0, 0, 0,
    SUM(openings)::numeric,
    0, 0, 0, 0,
    SUM(openings)::numeric,
    NULL
  FROM hr_vacancy
  GROUP BY department

  UNION ALL

  SELECT
    'leave_type',
    leave_type,
    CASE leave_type
      WHEN 'vacation' THEN 'Отпуск'
      WHEN 'sick' THEN 'Больничный'
      WHEN 'unpaid' THEN 'Без содержания'
      WHEN 'parental' THEN 'Декрет'
      ELSE leave_type
    END,
    COUNT(*)::numeric,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, NULL
  FROM hr_leave
  WHERE status IN ('approved', 'pending')
  GROUP BY leave_type

  UNION ALL

  SELECT
    'salary_band',
    salary_band,
    'Грейд ' || salary_band,
    COUNT(*)::numeric,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, NULL
  FROM hr_employee
  WHERE status IN ('active', 'probation', 'on_leave')
  GROUP BY salary_band
)
SELECT
  row_number() OVER (ORDER BY slice_kind, headcount DESC NULLS LAST, slice_label)::numeric AS id,
  slice_kind,
  slice_key,
  slice_label,
  ROUND(COALESCE(headcount, 0)::numeric, 1) AS headcount,
  ROUND(COALESCE(active_count, 0)::numeric, 1) AS active_count,
  ROUND(COALESCE(probation_count, 0)::numeric, 1) AS probation_count,
  ROUND(COALESCE(on_leave_count, 0)::numeric, 1) AS on_leave_count,
  ROUND(COALESCE(terminated_count, 0)::numeric, 1) AS terminated_count,
  ROUND(COALESCE(open_positions, 0)::numeric, 1) AS open_positions,
  ROUND(COALESCE(pending_leave, 0)::numeric, 1) AS pending_leave,
  ROUND(COALESCE(currently_away, 0)::numeric, 1) AS currently_away,
  ROUND(COALESCE(turnover_pct, 0)::numeric, 1) AS turnover_pct,
  ROUND(COALESCE(avg_tenure, 0)::numeric, 1) AS avg_tenure,
  ROUND(COALESCE(openings, 0)::numeric, 1) AS openings,
  extra_label
FROM u;

CREATE INDEX IF NOT EXISTS hr_employee_status_idx ON hr_employee (status);
CREATE INDEX IF NOT EXISTS hr_employee_dept_idx ON hr_employee (department);
CREATE INDEX IF NOT EXISTS hr_slice_kind_idx ON hr_slice (slice_kind);
