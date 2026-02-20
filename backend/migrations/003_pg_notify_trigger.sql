-- Migración 003: función + trigger para procesamiento asíncrono nativo de PostgreSQL
--
-- Flujo:
--   INSERT en applications
--     → trigger llama a notify_application_created()
--     → pg_notify publica en el canal 'bravo.application_created'
--     → PgListener en Go recibe la notificación
--     → PgListener publica el evento en RabbitMQ
--     → Workers existentes (risk_evaluator, auditor, notifier) lo procesan
--
-- Esto desacopla la API del broker: la API solo escribe en BD,
-- y la BD es quien dispara el procesamiento asíncrono.

CREATE OR REPLACE FUNCTION notify_application_created()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify(
        'bravo.application_created',
        json_build_object(
            'application_id', NEW.id,
            'user_id',        NEW.user_id,
            'country',        NEW.country,
            'status',         NEW.status,
            'created_at',     NEW.created_at
        )::text
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_application_created ON applications;

CREATE TRIGGER trg_application_created
    AFTER INSERT ON applications
    FOR EACH ROW
    EXECUTE FUNCTION notify_application_created();
