local project = import 'brewkit/project.libsonnet';

// TODO: appID поменять

local appIDs = [
    'inventory',
];

local proto = [
    'api/client/testinternal/testinternal.proto',
    'api/server/inventoryinternal/inventoryinternal.proto',
];

project.project(appIDs, proto)