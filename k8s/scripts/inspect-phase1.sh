#!/usr/bin/env bash
set -euo pipefail

# Phase 1 Inspection Helper
# Creates a test pod and provides inspection commands

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

IMAGE="${1:-cdk-erigon:k8s-test-phase1}"
NAMESPACE="phase1-inspection"
POD_NAME="inspect-$(date +%s)"

echo -e "${GREEN}Phase 1 Inspection Setup${NC}"
echo "Image: ${IMAGE}"
echo ""

# Create namespace
echo -e "${BLUE}Creating inspection namespace...${NC}"
kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f - > /dev/null 2>&1

# Create inspection pod
echo -e "${BLUE}Deploying inspection pod...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: ${POD_NAME}
  namespace: ${NAMESPACE}
  labels:
    app: phase1-inspection
spec:
  containers:
  - name: erigon
    image: ${IMAGE}
    imagePullPolicy: Never
    command: ["sleep", "infinity"]
    resources:
      requests:
        memory: "256Mi"
        cpu: "100m"
      limits:
        memory: "512Mi"
        cpu: "500m"
  restartPolicy: Never
EOF

echo -e "${BLUE}Waiting for pod to be ready...${NC}"
kubectl wait --for=condition=Ready pod/${POD_NAME} -n ${NAMESPACE} --timeout=60s

echo ""
echo -e "${GREEN}✓ Inspection pod ready!${NC}"
echo ""
echo -e "${YELLOW}=== INSPECTION COMMANDS ===${NC}"
echo ""
echo -e "${BLUE}# Check binary present:${NC}"
echo "kubectl exec -n ${NAMESPACE} ${POD_NAME} -- ls -lh /usr/local/bin/cdk-erigon"
echo ""
echo -e "${BLUE}# Test binary version:${NC}"
echo "kubectl exec -n ${NAMESPACE} ${POD_NAME} -- cdk-erigon --version"
echo ""
echo -e "${BLUE}# Check debug tools:${NC}"
echo "kubectl exec -n ${NAMESPACE} ${POD_NAME} -- which curl jq dig"
echo ""
echo -e "${BLUE}# Check config files:${NC}"
echo "kubectl exec -n ${NAMESPACE} ${POD_NAME} -- ls -lh /home/erigon/*.yaml"
echo ""
echo -e "${BLUE}# Check user and permissions:${NC}"
echo "kubectl exec -n ${NAMESPACE} ${POD_NAME} -- whoami"
echo "kubectl exec -n ${NAMESPACE} ${POD_NAME} -- id"
echo "kubectl exec -n ${NAMESPACE} ${POD_NAME} -- pwd"
echo ""
echo -e "${BLUE}# Interactive shell:${NC}"
echo "kubectl exec -it -n ${NAMESPACE} ${POD_NAME} -- /bin/sh"
echo ""
echo -e "${BLUE}# View pod details:${NC}"
echo "kubectl describe pod/${POD_NAME} -n ${NAMESPACE}"
echo ""
echo -e "${BLUE}# Check image locally:${NC}"
echo "docker run --rm ${IMAGE} cdk-erigon --version"
echo "docker run --rm --entrypoint /bin/sh ${IMAGE} -c 'ls -1 /usr/local/bin/cdk-erigon'"
echo ""
echo -e "${YELLOW}=== CLEANUP ===${NC}"
echo "kubectl delete namespace ${NAMESPACE}"
echo ""
