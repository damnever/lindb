/*
Licensed to LinDB under one or more contributor
license agreements. See the NOTICE file distributed with
this work for additional information regarding copyright
ownership. LinDB licenses this file to you under
the Apache License, Version 2.0 (the "License"); you may
not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
 
Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/
export enum Route {
  Overview = "/overview",
  StorageOverview = "/overview/storage",
  ConfigurationView = "/overview/configuration",
  Search = "/search",
  Explore = "/explore",
  MonitoringDashboard = "/monitoring/dashboard",
  MonitoringReplication = "/monitoring/replication",
  MonitoringRequest = "/monitoring/request",
  MonitoringLogs = "/monitoring/logs",
  MetadataStorage = "/metadata/storage",
  MetadataBroker = "/metadata/broker",
  MetadataStorageConfig = "/metadata/storage/configuration",
  MetadataBrokerConfig = "/metadata/broker/configuration",
  MetadataDatabase = "/metadata/database",
  MetadataDatabaseConfig = "/metadata/database/configuration",
  MetadataDatabaseLimits = "/metadata/database/limits",
  MetadataLogicDatabase = "/metadata/logic/database",
  MetadataLogicDatabaseConfig = "/metadata/logic/database/configuration",
  MetadataExplore = "/metadata/explore",
  MetadataMultipleIDC = "/metadata/multiple-idcs",
}

export enum ClusterType {
  Broker = "Broker",
  Storage = "Storage",
}
