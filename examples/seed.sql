-- Demo telemetry for GET /api/v1/gpus and /api/v1/gpus/{gpu_id}/telemetry
--
-- Time filters use processed_at_unix_nano (same instant as Go time.UnixNano() for RFC3339).
-- Values like 1000000001000000000 are NOT 2026 — they fall outside a 2026 query window.
-- Re-run safely after prior demo load:
DELETE FROM telemetry WHERE gpu_id IN ('gpu-001', 'gpu-002', 'gpu-003');

INSERT INTO public.telemetry (
  metric_name, gpu_id, device, uuid, model_name, host_name,
  value, labels_raw, processed_at_unix_nano, created_at
) VALUES
  ('gpu_utilization', 'gpu-001', 'cuda:0', '11111111-1111-1111-1111-111111111101', 'A100', 'host-a', 0.42, '{"zone":"us"}',  1777622400000000000, '2026-05-01 08:00:00+00'),
  ('gpu_utilization', 'gpu-001', 'cuda:0', '11111111-1111-1111-1111-111111111102', 'A100', 'host-a', 0.71, '{"zone":"us"}',  1777631400000000000, '2026-05-01 10:30:00+00'),
  ('gpu_memory_used', 'gpu-001', 'cuda:0', '11111111-1111-1111-1111-111111111103', 'A100', 'host-a', 38.2, '{"unit":"GiB"}', 1777644900000000000, '2026-05-01 14:15:00+00'),
  ('gpu_utilization', 'gpu-001', 'cuda:0', '11111111-1111-1111-1111-111111111104', 'A100', 'host-a', 0.88, '{"zone":"us"}',  1777712400000000000, '2026-05-02 09:00:00+00'),
  ('gpu_utilization', 'gpu-001', 'cuda:0', '11111111-1111-1111-1111-111111111105', 'A100', 'host-a', 0.15, '{"zone":"us"}',  1777765500000000000, '2026-05-02 23:45:00+00'),

  ('gpu_utilization', 'gpu-002', 'cuda:1', '22222222-2222-2222-2222-222222222201', 'H100', 'host-b', 0.55, '{"zone":"eu"}',  1777618800000000000, '2026-05-01 07:00:00+00'),
  ('gpu_utilization', 'gpu-002', 'cuda:1', '22222222-2222-2222-2222-222222222202', 'H100', 'host-b', 0.92, '{"zone":"eu"}',  1777723200000000000, '2026-05-02 12:00:00+00'),
  ('gpu_memory_used', 'gpu-002', 'cuda:1', '22222222-2222-2222-2222-222222222203', 'H100', 'host-b', 72.0, '{"unit":"GiB"}', 1777789800000000000, '2026-05-03 06:30:00+00'),

  ('gpu_utilization', 'gpu-003', 'cuda:0', '33333333-3333-3333-3333-333333333301', 'L4',  'host-c', 0.33, '{"zone":"ap"}', 1777766400000000000, '2026-05-03 00:00:00+00'),
  ('gpu_utilization', 'gpu-003', 'cuda:0', '33333333-3333-3333-3333-333333333302', 'L4',  'host-c', 0.66, '{"zone":"ap"}', 1777809600000000000, '2026-05-03 12:00:00+00'),
  ('gpu_utilization', 'gpu-003', 'cuda:0', '33333333-3333-3333-3333-333333333303', 'L4',  'host-c', 0.99, '{"zone":"ap"}', 1777833900000000000, '2026-05-03 18:45:00+00'),

  ('gpu_utilization', 'gpu-001', 'cuda:0', '11111111-1111-1111-1111-111111111199', 'A100', 'host-a', 0.50, '{"edge":true}', 1777852740000000000, '2026-05-03 23:59:00+00');
