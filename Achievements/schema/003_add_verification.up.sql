-- F4: экспертная верификация портфолио.
-- verification_status: 1=DRAFT, 2=PENDING, 3=APPROVED, 4=REJECTED (см. proto VerificationStatus).
ALTER TABLE achievements
    ADD COLUMN verification_status SMALLINT NOT NULL DEFAULT 1,
    ADD COLUMN reviewed_by VARCHAR(36),
    ADD COLUMN reviewed_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN review_comment TEXT;

-- Индекс для очереди эксперта (PENDING-выборка).
CREATE INDEX idx_achievements_verification_pending
    ON achievements(verification_status, created_at)
    WHERE deleted_at IS NULL AND verification_status = 2;
