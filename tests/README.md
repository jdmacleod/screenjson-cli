# Test Files

Example conversions generated from `His Girl Friday (1940)` screenplay.

## Generated Files

| File | Description | Size |
|------|-------------|------|
| `895af360-50aa-4d89-8fa4-a39e92a01297.json` | FDX → ScreenJSON | 2.0M |
| `9fea5fc5-f8f9-4c59-85d2-9ec5dcd452c1.json` | Fountain → ScreenJSON | 2.0M |
| `69beb802-c6db-4653-9cb8-1dbd36ad6acf.json` | FadeIn → ScreenJSON | 2.0M |
| `69154844-3f50-4f14-9dbb-c6fa088325e3.pdf` | ScreenJSON → PDF | 222K |
| `b639c0bb-eff7-4b50-896d-923f1fc425d8.fdx` | ScreenJSON → FDX | 648K |
| `bc9e038b-3ec5-4091-9095-df9c973bed47.fountain` | ScreenJSON → Fountain | 176K |
| `5b136080-68e9-4cfa-b44c-e4312d484da0.fadein` | ScreenJSON → FadeIn | 92K |
| `d5bc980c-075f-4f37-98dc-db059663f4bc.yaml` | Fountain → ScreenYAML | 5.0M |
| `3c5e006f-4503-4e0b-b23e-3bff8e617084_encrypted.json` | Encrypted ScreenJSON | 2.3M |
| `bb6ef9b6-e0c6-426f-8fd2-30e8a59e69a5_decrypted.json` | Decrypted ScreenJSON | 2.0M |

## Commands Used

```bash
# Import conversions
screenjson convert -o <uuid>.json examples/His_Girl_Friday_1940_Screenplay.fdx
screenjson convert -o <uuid>.json examples/His_Girl_Friday_1940_Screenplay.fountain
screenjson convert -o <uuid>.json examples/His_Girl_Friday_1940_Screenplay.fadein

# Export conversions
screenjson export -f pdf -o <uuid>.pdf <input>.json
screenjson export -f fdx -o <uuid>.fdx <input>.json
screenjson export -f fountain -o <uuid>.fountain <input>.json
screenjson export -f fadein -o <uuid>.fadein <input>.json

# YAML output
screenjson convert --yaml -o <uuid>.yaml examples/His_Girl_Friday_1940_Screenplay.fountain

# Encryption/Decryption
screenjson encrypt --key "TopSecretKey2026!" -o <uuid>_encrypted.json <input>.json
screenjson decrypt --key "TopSecretKey2026!" -o <uuid>_decrypted.json <encrypted>.json
```

## Validation

All ScreenJSON files pass validation:

```bash
screenjson validate <file>.json
# ✓ Document is valid
```
