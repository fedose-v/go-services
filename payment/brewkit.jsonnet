local project = import 'brewkit/project.libsonnet';

local appIDs = [
    'payment',
];

local proto = [
    'api/server/paymentpublicapi/paymentpublicapi.proto',
];

project.project(appIDs, proto)