-- Remove only the attribution written by the SHUZ-152 confirmation migration.
UPDATE project
SET created_by = NULL
WHERE workspace_id = '5d54d458-153b-4612-a1a4-e3346de2bcb2'::uuid
  AND (
      (
          created_by = 'b4a22a55-a3aa-499e-b33e-26de4b8fc65a'::uuid
          AND id IN (
              '7181cff2-8d1f-4cb4-9cae-c4959cf4c697'::uuid,
              'e4f36ec8-3c67-47d5-a2c2-08f458efb1af'::uuid
          )
      )
      OR (
          created_by = 'a01b5b0c-0491-4186-a8c4-f33f25f834c9'::uuid
          AND id IN (
              '0bf8cd8a-4604-4a1a-a754-03c1955e5cb8'::uuid,
              'ef04ece8-8c18-4c34-b700-9f9cfff3124b'::uuid,
              'a004c263-8430-453c-bc5f-76400ecfc6a6'::uuid,
              '2b0ae02a-7b8b-432a-b93f-bf4f047ab6a9'::uuid,
              'afcad442-1592-4bca-b6cd-f4edf5e9906f'::uuid,
              '66dd9524-df97-4e2f-a946-fe0a39f36a89'::uuid,
              '123bbedd-461f-4c1c-b50c-6c7842e09266'::uuid,
              '1c5c03f6-701c-4c7a-80bb-6158bcf969c3'::uuid,
              '64179265-069e-4b01-8343-921ac2c07134'::uuid,
              '144359fb-b157-4b46-9cfc-61d179698e67'::uuid,
              'dd684e92-b204-463b-809e-854f79c89ca2'::uuid,
              'd1e511aa-cf63-4bc0-8ef1-2eddc9ee888e'::uuid
          )
      )
  );
