# tool-ecg-backend
Tools@ECG Backend

## Setup Empty Standplan
SQL Script:
```
BEGIN;

-- -------------------------------------------------------------
-- 1. Tage anlegen (3 Wochenenden: Fr, Sa, So)
--    Weekend IDs: 1, 2, 3
-- -------------------------------------------------------------
INSERT INTO days (weekend_id, date) VALUES
-- Wochenende 1
(1, '2026-11-27'), -- Fr
(1, '2026-11-28'), -- Sa
(1, '2026-11-29'), -- So
-- Wochenende 2
(2, '2026-12-04'), -- Fr
(2, '2026-12-05'), -- Sa
(2, '2026-12-06'), -- So
-- Wochenende 3
(3, '2026-12-11'), -- Fr
(3, '2026-12-12'), -- Sa
(3, '2026-12-13'); -- So

-- -------------------------------------------------------------
-- Helper / CTEs zur Erstellung aller Schichten für alle 3 Wochenenden
-- -------------------------------------------------------------
WITH day_map AS (
    SELECT id, weekend_id, EXTRACT(ISODOW FROM date) AS dow
    FROM days
),
shift_templates AS (
    -- Freitag (dow = 5)
    SELECT 5 AS dow, 'Aufbauen' AS title, TIME '11:30' AS start_time, TIME '12:00' AS end_time, 1 AS slots UNION ALL
    SELECT 5, '1. Schicht', TIME '11:30', TIME '14:30', 3 UNION ALL
    SELECT 5, '2. Schicht', TIME '14:30', TIME '17:00', 3 UNION ALL
    SELECT 5, '3. Schicht', TIME '17:00', TIME '20:00', 3 UNION ALL
    SELECT 5, '4. Schicht', TIME '20:00', TIME '22:30', 3 UNION ALL
    SELECT 5, 'Aufräumen',  TIME '21:30', TIME '22:30', 1 UNION ALL
    SELECT 5, 'Teigbeauftragter', TIME '11:00', TIME '22:30', 1 UNION ALL

    -- Samstag (dow = 6)
    SELECT 6, 'Aufbauen', TIME '10:30', TIME '11:00', 1 UNION ALL
    SELECT 6, '1. Schicht', TIME '10:30', TIME '13:30', 3 UNION ALL
    SELECT 6, '2. Schicht', TIME '13:30', TIME '16:30', 3 UNION ALL
    SELECT 6, '3. Schicht', TIME '16:30', TIME '19:30', 3 UNION ALL
    SELECT 6, '4. Schicht', TIME '19:30', TIME '22:30', 3 UNION ALL
    SELECT 6, 'Aufräumen',  TIME '21:30', TIME '22:30', 1 UNION ALL
    SELECT 6, 'Teigbeauftragter', TIME '10:00', TIME '22:30', 1 UNION ALL

    -- Sonntag (dow = 7)
    SELECT 7, 'Aufbauen', TIME '10:30', TIME '11:00', 1 UNION ALL
    SELECT 7, '1. Schicht', TIME '10:30', TIME '13:30', 3 UNION ALL
    SELECT 7, '2. Schicht', TIME '13:30', TIME '16:30', 3 UNION ALL
    SELECT 7, '3. Schicht', TIME '16:30', TIME '19:30', 3 UNION ALL
    SELECT 7, '4. Schicht', TIME '19:30', TIME '21:00', 3 UNION ALL
    SELECT 7, 'Aufräumen',  TIME '20:00', TIME '21:00', 1 UNION ALL
    SELECT 7, 'Teigbeauftragter', TIME '10:00', TIME '21:00', 1
)
INSERT INTO shifts (day_id, title, start_time, end_time, required_slots, description)
SELECT 
    dm.id AS day_id,
    st.title,
    st.start_time,
    st.end_time,
    st.slots AS required_slots,
    CASE 
        WHEN st.title = 'Teigbeauftragter' THEN 'Zuständig für Nachschub & Teigvorbereitung'
        WHEN st.title IN ('Aufbauen', 'Aufräumen') THEN 'Hütten-Service & Logistik'
        ELSE 'Verkauf & Waffelbacken'
    END AS description
FROM day_map dm
JOIN shift_templates st ON dm.dow = st.dow
ORDER BY dm.id, st.start_time;

COMMIT;
```
