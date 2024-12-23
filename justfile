shell:
	nix --experimental-features 'nix-command flakes' develop

dev:
	cd example && air

client:
	cd example/client && npm i && npm run dev
