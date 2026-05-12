include ../../../app.mk

# override openapi target for game-admin
openapi:
	@cd ../../../api && \
	buf generate --template buf.game-admin.openapi.gen.yaml
