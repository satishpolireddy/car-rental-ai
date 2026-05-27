# AKS Cluster — Kubernetes on Azure for horizontal auto-scaling
resource "azurerm_kubernetes_cluster" "main" {
  name                = "${var.prefix}-aks"
  location            = var.location
  resource_group_name = var.resource_group_name
  dns_prefix          = "${var.prefix}-aks"
  kubernetes_version  = var.kubernetes_version

  default_node_pool {
    name                = "system"
    node_count          = var.system_node_count
    vm_size             = var.system_vm_size
    os_disk_size_gb     = 50
    type                = "VirtualMachineScaleSets"
    enable_auto_scaling = true
    min_count           = 1
    max_count           = var.system_max_nodes
  }

  # App node pool — auto-scales based on load
  dynamic "auto_scaler_profile" {
    for_each = [1]
    content {
      scale_down_delay_after_add  = "5m"
      scale_down_unneeded         = "2m"
      max_graceful_termination_sec = "600"
    }
  }

  identity {
    type = "SystemAssigned"
  }

  network_profile {
    network_plugin    = "azure"
    load_balancer_sku = "standard"
  }

  tags = var.tags
}

# Separate node pool for application workloads
resource "azurerm_kubernetes_cluster_node_pool" "app" {
  name                  = "app"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.main.id
  vm_size               = var.app_vm_size
  node_count            = var.app_node_count
  enable_auto_scaling   = true
  min_count             = var.app_min_nodes
  max_count             = var.app_max_nodes
  os_disk_size_gb       = 50
  node_labels = {
    "role" = "app"
  }
  tags = var.tags
}

output "kube_config" {
  value     = azurerm_kubernetes_cluster.main.kube_config_raw
  sensitive = true
}

output "cluster_name" {
  value = azurerm_kubernetes_cluster.main.name
}
