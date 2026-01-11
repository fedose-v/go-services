local project = import 'brewkit/project.libsonnet';

local appIDs = [
    'inventory',
];

local proto = [
    'api/server/inventorypublicapi/inventorypublicapi.proto',
];

project.project(appIDs, proto)