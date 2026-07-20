#!/usr/bin/env bash
set -euo pipefail

echo ""
echo "=============================================="
echo " Instalando GitHub Actions Runner (cluster unificado)"
echo "=============================================="

# --- Detectar SO ---
if [ -f /etc/alpine-release ]; then
    OS="alpine"
elif [ -f /etc/debian_version ]; then
    OS="debian"
elif [ -f /etc/redhat-release ]; then
    OS="rhel"
else
    OS="unknown"
fi
echo " Sistema detectado: $OS"
echo ""

# --- Pedir datos ---
read -rp "URL del repositorio (ej: https://github.com/usuario/repo): " REPO_URL
read -rp "Token del runner (de la pagina de GitHub): " RUNNER_TOKEN

RUNNER_DIR="$HOME/actions-runner"

# =============================================================================
# 1. Instalar dependencias segun SO
# =============================================================================
echo "[1/6] Instalando dependencias..."

install_deps_alpine() {
    echo "    (Alpine detectado - instalando con apk)"
    sudo apk add --no-cache \
        curl tar git \
        libstdc++ icu-libs icu gcompat \
        docker kubectl 2>/dev/null || true

    if ! command -v kubectl &>/dev/null; then
        echo "    Instalando kubectl manualmente..."
        curl -LO "https://dl.k8s.io/release/v1.28.0/bin/linux/amd64/kubectl"
        chmod +x kubectl
        sudo mv kubectl /usr/local/bin/
    fi
}

install_deps_debian() {
    echo "    (Debian/Ubuntu detectado - instalando con apt)"
    sudo apt-get update -qq
    sudo apt-get install -y -qq curl tar git 2>/dev/null || true
    if ! command -v kubectl &>/dev/null; then
        curl -LO "https://dl.k8s.io/release/v1.28.0/bin/linux/amd64/kubectl"
        chmod +x kubectl
        sudo mv kubectl /usr/local/bin/
    fi
}

install_deps_rhel() {
    echo "    (RHEL/CentOS detectado - instalando con dnf)"
    sudo dnf install -y curl tar git 2>/dev/null || true
    if ! command -v kubectl &>/dev/null; then
        curl -LO "https://dl.k8s.io/release/v1.28.0/bin/linux/amd64/kubectl"
        chmod +x kubectl
        sudo mv kubectl /usr/local/bin/
    fi
}

case "$OS" in
    alpine) install_deps_alpine ;;
    debian) install_deps_debian ;;
    rhel)   install_deps_rhel ;;
    *)      echo "[!] SO no reconocido. Instala curl, tar, git, kubectl manualmente." ;;
esac

# =============================================================================
# 2. Verificar kubectl
# =============================================================================
echo "[2/6] Verificando kubectl..."
if command -v kubectl &>/dev/null; then
    echo "    kubectl: $(kubectl version --client --short 2>/dev/null || kubectl version --client 2>/dev/null || echo 'OK')"
    kubectl cluster-info 2>/dev/null && echo "    Cluster: CONECTADO" || echo "    [!] Cluster no accesible. Configura kubeconfig primero."
else
    echo "    [!] ERROR: kubectl no se pudo instalar."
    exit 1
fi

# =============================================================================
# 3. Verificar Docker
# =============================================================================
echo "[3/6] Verificando Docker..."
if command -v docker &>/dev/null; then
    echo "    Docker: $(docker --version)"
else
    echo "    [!] ADVERTENCIA: Docker no encontrado."
    case "$OS" in
        alpine) echo "    Instalar: sudo apk add docker" ;;
        debian) echo "    Instalar: sudo apt-get install -y docker.io" ;;
        rhel)   echo "    Instalar: sudo dnf install -y docker" ;;
    esac
fi

# =============================================================================
# 4. Verificar kubeconfig
# =============================================================================
echo "[4/6] Verificando kubeconfig..."
if [ -f "$HOME/.kube/config" ]; then
    echo "    Kubeconfig: ENCONTRADO"
else
    echo "    [!] ADVERTENCIA: ~/.kube/config no existe."
    echo "    Copia el kubeconfig de VM1:"
    echo "      mkdir -p ~/.kube"
    echo "      sudo cat /etc/rancher/k3s/k3s.yaml | sed 's/127.0.0.1/146.83.102.21/g' > ~/.kube/config"
fi

# =============================================================================
# 5. Descargar e instalar runner
# =============================================================================
echo "[5/6] Descargando runner de GitHub..."

if [ -d "$RUNNER_DIR" ]; then
    echo "    Directorio existente, eliminando..."
    rm -rf "$RUNNER_DIR"
fi
mkdir -p "$RUNNER_DIR"
cd "$RUNNER_DIR"

curl -o actions-runner.tar.gz -L \
    https://github.com/actions/runner/releases/download/v2.322.0/actions-runner-linux-x64-2.322.0.tar.gz
echo "b13b784808359f31bc79b08a191f5f83757852957dd8fe3dbfcc38202ccf5768  actions-runner.tar.gz" | shasum -a 256 -c -
tar xzf actions-runner.tar.gz
rm actions-runner.tar.gz

echo "[6/6] Configurando runner..."
./config.sh --url "$REPO_URL" --token "$RUNNER_TOKEN" \
    --name "sat-$(hostname)" \
    --labels "sat" \
    --work "_work" \
    --unattended \
    --replace

# =============================================================================
# 6. Instalar como servicio
# =============================================================================
echo ""
echo "=============================================="
echo " Instalando como servicio..."

if [ "$OS" = "alpine" ]; then
    echo " (Alpine usa OpenRC en vez de systemd)"
    echo ""
    echo " Ejecuta el runner en background con:"
    echo "   cd $RUNNER_DIR && nohup ./run.sh &"
    echo ""
    echo " O crea un script de inicio en /etc/local.d/:"
    echo "   echo 'cd $RUNNER_DIR && nohup ./run.sh &' | sudo tee /etc/local.d/runner.start"
    echo "   sudo chmod +x /etc/local.d/runner.start"
    echo "   sudo rc-update add local"
else
    sudo ./svc.sh install
    sudo ./svc.sh start
    echo ""
    echo " Verificar estado: sudo ./svc.sh status"
fi

echo ""
echo "=============================================="
echo " Runner listo (cluster unificado)"
echo " Etiqueta:          sat"
echo " Directorio:        $RUNNER_DIR"
echo " Logs:              tail -f $RUNNER_DIR/_diag/*.log"
echo ""
echo " IMPORTANTE: en GitHub, el runner debe aparecer como 'Idle'"
echo " en Settings -> Actions -> Runners"
echo "=============================================="
