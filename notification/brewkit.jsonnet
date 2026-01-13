local project = import 'brewkit/project.libsonnet';

local appIDs = [
    'notification',
];

local proto = [
    'api/server/notificationinternal/notificationinternal.proto',
];

project.project(appIDs, proto)