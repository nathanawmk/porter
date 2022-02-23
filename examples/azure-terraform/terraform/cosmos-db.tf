resource "azurerm_resource_group" "rg" {
  name     = var.resource_group_name
  location = var.resource_group_location
  tags = {
    yor_trace = "4a7916c9-aa89-4795-b845-11a5f0a7c24b"
  }
}

resource "azurerm_cosmosdb_account" "db" {
  name                = "porterform-cosmos-db"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  offer_type          = "Standard"
  kind                = "MongoDB"

  enable_automatic_failover = false

  consistency_policy {
    consistency_level       = "BoundedStaleness"
    max_interval_in_seconds = 301
    max_staleness_prefix    = 100001
  }

  geo_location {
    location          = var.failover_location
    failover_priority = 1
  }

  geo_location {
    prefix            = "porterform-${azurerm_resource_group.rg.location}"
    location          = azurerm_resource_group.rg.location
    failover_priority = 0
  }
  tags = {
    yor_trace = "e5f65ca8-3922-4fbf-97b9-d3a6b338a1b3"
  }
}

resource "azurerm_cosmosdb_mongo_database" "db" {
  name                = var.database_name
  resource_group_name = azurerm_cosmosdb_account.db.resource_group_name
  account_name        = azurerm_cosmosdb_account.db.name
}