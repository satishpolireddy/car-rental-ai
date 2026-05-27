terraform {
  required_version = ">= 1.7"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.100"
    }
  }
  # Remote state — ensures team collaboration and prevents drift
  backend "azurerm" {
    resource_group_name  = "tfstate-rg"
    storage_account_name = "carrentalaitfstate"
    container_name       = "tfstate"
    key                  = "prod.terraform.tfstate"
  }
}

provider "azurerm" {
  features {}
}

locals {
  prefix   = "carrental-prod"
  location = "eastus"
  tags = {
    Environment = "production"
    Project     = "car-rental-ai"
    ManagedBy   = "terraform"
  }
}

resource "azurerm_resource_group" "main" {
  name     = "${local.prefix}-rg"
  location = local.location
  tags     = local.tags
}

# SQL Server — managed Azure SQL for zero-maintenance scaling
resource "azurerm_mssql_server" "main" {
  name                         = "${local.prefix}-sql"
  resource_group_name          = azurerm_resource_group.main.name
  location                     = local.location
  version                      = "12.0"
  administrator_login          = var.db_admin_user
  administrator_login_password = var.db_admin_password
  tags                         = local.tags
}

resource "azurerm_mssql_database" "main" {
  name        = "carrental"
  server_id   = azurerm_mssql_server.main.id
  collation   = "SQL_Latin1_General_CP1_CI_AS"
  sku_name    = "GP_Gen5_4"   # General Purpose — auto-scalable
  max_size_gb = 100
  tags        = local.tags
}

# Redis Cache — for AI response caching, session state
resource "azurerm_redis_cache" "main" {
  name                = "${local.prefix}-redis"
  resource_group_name = azurerm_resource_group.main.name
  location            = local.location
  capacity            = 1
  family              = "C"
  sku_name            = "Standard"   # Standard = replicated, survives node failure
  enable_non_ssl_port = false
  minimum_tls_version = "1.2"
  tags                = local.tags
}

# Azure OpenAI
resource "azurerm_cognitive_account" "openai" {
  name                = "${local.prefix}-openai"
  resource_group_name = azurerm_resource_group.main.name
  location            = "eastus"
  kind                = "OpenAI"
  sku_name            = "S0"
  tags                = local.tags
}

resource "azurerm_cognitive_deployment" "gpt4o" {
  name                 = "gpt-4o"
  cognitive_account_id = azurerm_cognitive_account.openai.id
  model {
    format  = "OpenAI"
    name    = "gpt-4o"
    version = "2024-05-13"
  }
  scale {
    type     = "Standard"
    capacity = 30  # tokens per minute (thousands)
  }
}

# AKS Cluster
module "aks" {
  source              = "../../modules/aks"
  prefix              = local.prefix
  location            = local.location
  resource_group_name = azurerm_resource_group.main.name
  kubernetes_version  = "1.29"
  system_node_count   = 2
  system_vm_size      = "Standard_D2s_v3"
  system_max_nodes    = 5
  app_vm_size         = "Standard_D4s_v3"
  app_node_count      = 2
  app_min_nodes       = 2
  app_max_nodes       = 10   # scales to 10 nodes under load
  tags                = local.tags
}

# Container Registry
resource "azurerm_container_registry" "main" {
  name                = replace("${local.prefix}acr", "-", "")
  resource_group_name = azurerm_resource_group.main.name
  location            = local.location
  sku                 = "Premium"
  admin_enabled       = false
  tags                = local.tags
}

output "aks_cluster_name" { value = module.aks.cluster_name }
output "openai_endpoint"  { value = azurerm_cognitive_account.openai.endpoint }
output "redis_hostname"   { value = azurerm_redis_cache.main.hostname }
output "sql_server_fqdn"  { value = azurerm_mssql_server.main.fully_qualified_domain_name }
