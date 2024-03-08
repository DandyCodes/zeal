# =============
# CUSTOMISATION
# =============
sed -i '0,/ZSH_THEME="devcontainers"/s//ZSH_THEME="robbyrussell"/' ~/.zshrc;
echo "neofetch" >> ~/.zshrc;
mkdir -p ~/.config/helix && echo "theme=\"dark_plus\"" >> ~/.config/helix/config.toml;

# ==========
# SSH CONFIG
# ==========
sudo cp -r /root/.ssh ~;
sudo chown -R vscode:vscode ~/.ssh;
sudo chmod 700 -R ~/.ssh;