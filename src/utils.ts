import { QueryModelMetricsAggregation } from './types';

export const nameOptions: string[] = [
  'httpRequests',
  'httpRequests_overview_bytes',
  'httpRequests_overview_cachedBytes',
  'httpRequests_overview_cachedRequests',
  'httpRequests_overview_pageViews',
  'httpRequests_overview_requests',
  'httpRequests_overview_visits',
  'httpRequests_overview_originResponseDurationMs',
  'httpRequests_edgeDnsResponseTimeMs',
  'httpRequests_edgeRequestBytes',
  'httpRequests_edgeResponseBytes',
  'httpRequests_edgeTimeToFirstByteMs',
  'httpRequests_originResponseDurationMs',
  'httpRequests_visits',
  'firewallEvents',
  'firewallEvents_events',
  'workersInvocations_clientDisconnects',
  'workersInvocations_cpuTime',
  'workersInvocations_duration',
  'workersInvocations_errors',
  'workersInvocations_requestDuration',
  'workersInvocations_requests',
  'workersInvocations_responseBodySize',
  'workersInvocations_subrequests',
  'workersInvocations_wallTime',
  'workersLogs',
];

// isAccountScopedMetric returns true when the provided metric is scoped to a
// Cloudflare account instead of a zone, so that the query requires an account
// instead of a zone.
export function isAccountScopedMetric(name: string): boolean {
  return name === 'workersLogs' || name.startsWith('workersInvocations_');
}

export function getAggregationOptions(
  name: string,
): QueryModelMetricsAggregation[] | undefined {
  if (
    [
      'httpRequests_overview_bytes',
      'httpRequests_overview_cachedBytes',
      'httpRequests_overview_cachedRequests',
      'httpRequests_overview_pageViews',
      'httpRequests_overview_requests',
      'httpRequests_overview_visits',
    ].includes(name)
  ) {
    return ['sum'];
  }

  if (['httpRequests_overview_originResponseDurationMs'].includes(name)) {
    return ['avg'];
  }

  if (
    [
      'httpRequests_edgeDnsResponseTimeMs',
      'httpRequests_edgeTimeToFirstByteMs',
      'httpRequests_originResponseDurationMs',
    ].includes(name)
  ) {
    return ['avg', 'sum', 'count'];
  }

  if (
    [
      'httpRequests_edgeRequestBytes',
      'httpRequests_edgeResponseBytes',
      'httpRequests_visits',
    ].includes(name)
  ) {
    return ['sum', 'count'];
  }

  if (['firewallEvents_events'].includes(name)) {
    return ['count'];
  }

  if (
    [
      'workersInvocations_clientDisconnects',
      'workersInvocations_errors',
      'workersInvocations_requests',
      'workersInvocations_subrequests',
    ].includes(name)
  ) {
    return ['sum'];
  }

  if (
    [
      'workersInvocations_duration',
      'workersInvocations_responseBodySize',
      'workersInvocations_wallTime',
    ].includes(name)
  ) {
    return ['sum', 'P50', 'P75', 'P90', 'P99', 'P999'];
  }

  if (
    [
      'workersInvocations_cpuTime',
      'workersInvocations_requestDuration',
    ].includes(name)
  ) {
    return ['P50', 'P75', 'P90', 'P99', 'P999'];
  }

  return undefined;
}

export function getFiltersOptions(name: string): string[] {
  if (name.startsWith('httpRequests_overview_')) {
    return filtersOptions['httpRequestsOverview'];
  }

  if (name.startsWith('httpRequests')) {
    return filtersOptions['httpRequests'];
  }

  if (name.startsWith('firewallEvents')) {
    return filtersOptions['firewallEvents'];
  }

  if (name.startsWith('workersInvocations_')) {
    return filtersOptions['workersInvocations'];
  }

  if (name === 'workersLogs') {
    return filtersOptions['workersLogs'];
  }

  return [];
}

const filtersOptions: Record<string, string[]> = {
  httpRequestsOverview: [
    '-',
    'clientCountryName',
    'clientRequestHTTPProtocol',
    'clientSSLProtocol',
    'edgeResponseContentTypeName',
    'edgeResponseStatus',
    'httpApplicationVersion',
    'rayName',
    'userAgentBrowser',
    'zoneVersion',
  ],
  httpRequests: [
    '-',
    'botManagementDecision',
    'cacheStatus',
    'clientASNDescription',
    'clientAsn',
    'clientCountryName',
    'clientDeviceType',
    'clientIP',
    'clientRefererHost',
    'clientRequestHTTPHost',
    'clientRequestHTTPMethodName',
    'clientRequestHTTPProtocol',
    'clientRequestPath',
    'clientRequestQuery',
    'clientRequestReferer',
    'clientRequestScheme',
    'clientSSLProtocol',
    'coloCode',
    'edgeDnsResponseTimeMs',
    'edgeResponseContentTypeName',
    'edgeResponseStatus',
    'edgeTimeToFirstByteMs',
    'originASN',
    'originASNDescription',
    'originIP',
    'originResponseDurationMs',
    'originResponseStatus',
    'rayName',
    'requestSource',
    'securityAction',
    'upperTierColoName',
    'userAgent',
    'userAgentBrowser',
    'userAgentOS',
    'verifiedBotCategory',
    'wafAttackScore',
    'wafAttackScoreClass',
    'wafRceAttackScore',
    'wafSqliAttackScore',
    'wafXssAttackScore',
  ],
  firewallEvents: [
    '-',
    'action',
    'clientASNDescription',
    'clientAsn',
    'clientCountryName',
    'clientIP',
    'clientIPClass',
    'clientRefererHost',
    'clientRefererPath',
    'clientRefererQuery',
    'clientRefererScheme',
    'clientRequestHTTPHost',
    'clientRequestHTTPMethodName',
    'clientRequestHTTPProtocol',
    'clientRequestPath',
    'clientRequestQuery',
    'clientRequestScheme',
    'description',
    'edgeColoName',
    'edgeResponseStatus',
    'kind',
    'matchIndex',
    'originResponseStatus',
    'originatorRayName',
    'rayName',
    'ref',
    'ruleId',
    'rulesetId',
    'source',
    'userAgent',
    'verifiedBotCategory',
    'wafAttackScore',
    'wafAttackScoreClass',
    'wafRceAttackScore',
    'wafSqliAttackScore',
    'wafXssAttackScore',
  ],
  workersInvocations: [
    '-',
    'coloCode',
    'dispatchNamespaceName',
    'environmentName',
    'isDispatcher',
    'scriptName',
    'scriptTag',
    'scriptVersion',
    'status',
    'usageModel',
  ],
  workersLogs: [
    '-',
    '$metadata.error',
    '$metadata.level',
    '$metadata.message',
    '$metadata.requestId',
    '$metadata.service',
    '$metadata.trigger',
    '$workers.entrypoint',
    '$workers.event.request.method',
    '$workers.event.request.url',
    '$workers.event.response.status',
    '$workers.eventType',
    '$workers.executionModel',
    '$workers.outcome',
    '$workers.scriptName',
  ],
};

export const getDimensionsOptions = (name: string): string[] => {
  if (name.startsWith('httpRequests_overview_')) {
    return dimensionsOptions['httpRequestsOverview'];
  }

  if (name.startsWith('httpRequests_')) {
    return dimensionsOptions['httpRequests'];
  }

  if (name.startsWith('firewallEvents')) {
    return dimensionsOptions['firewallEvents'];
  }

  if (name.startsWith('workersInvocations_')) {
    return dimensionsOptions['workersInvocations'];
  }

  return [];
};

const dimensionsOptions: Record<string, string[]> = {
  httpRequestsOverview: [
    'clientCountryName',
    'clientRequestHTTPProtocol',
    'clientSSLProtocol',
    'date',
    'datetime',
    'datetimeFifteenMinutes',
    'datetimeFiveMinutes',
    'datetimeHour',
    'datetimeMinute',
    'edgeResponseContentTypeName',
    'edgeResponseStatus',
    'httpApplicationVersion',
    'userAgentBrowser',
    'zoneVersion',
  ],
  httpRequests: [
    'botManagementDecision',
    'cacheStatus',
    'clientASNDescription',
    'clientAsn',
    'clientCountryName',
    'clientDeviceType',
    'clientIP',
    'clientRefererHost',
    'clientRequestHTTPHost',
    'clientRequestHTTPMethodName',
    'clientRequestHTTPProtocol',
    'clientRequestPath',
    'clientRequestQuery',
    'clientRequestReferer',
    'clientRequestScheme',
    'clientSSLProtocol',
    'coloCode',
    'date',
    'datetime',
    'datetimeFifteenMinutes',
    'datetimeFiveMinutes',
    'datetimeHour',
    'datetimeMinute',
    'edgeDnsResponseTimeMs',
    'edgeResponseContentTypeName',
    'edgeResponseStatus',
    'edgeTimeToFirstByteMs',
    'originASN',
    'originASNDescription',
    'originIP',
    'originResponseDurationMs',
    'originResponseStatus',
    'requestSource',
    'securityAction',
    'upperTierColoName',
    'userAgent',
    'userAgentBrowser',
    'userAgentOS',
    'verifiedBotCategory',
    'wafAttackScore',
    'wafAttackScoreClass',
    'wafRceAttackScore',
    'wafSqliAttackScore',
    'wafXssAttackScore',
  ],
  firewallEvents: [
    'action',
    'clientASNDescription',
    'clientAsn',
    'clientCountryName',
    'clientIP',
    'clientIPClass',
    'clientRefererHost',
    'clientRefererPath',
    'clientRefererQuery',
    'clientRefererScheme',
    'clientRequestHTTPHost',
    'clientRequestHTTPMethodName',
    'clientRequestHTTPProtocol',
    'clientRequestPath',
    'clientRequestQuery',
    'clientRequestScheme',
    'date',
    'datetime',
    'datetimeFifteenMinutes',
    'datetimeFiveMinutes',
    'datetimeHour',
    'datetimeMinute',
    'description',
    'edgeColoName',
    'edgeResponseStatus',
    'kind',
    'matchIndex',
    'originResponseStatus',
    'originatorRayName',
    'rayName',
    'ref',
    'ruleId',
    'rulesetId',
    'source',
    'userAgent',
    'verifiedBotCategory',
    'wafAttackScore',
    'wafAttackScoreClass',
    'wafRceAttackScore',
    'wafSqliAttackScore',
    'wafXssAttackScore',
  ],
  workersInvocations: [
    'coloCode',
    'date',
    'datetime',
    'datetimeFifteenMinutes',
    'datetimeFiveMinutes',
    'datetimeHour',
    'datetimeMinute',
    'datetimeSixHours',
    'dispatchNamespaceName',
    'environmentName',
    'isDispatcher',
    'scriptName',
    'scriptTag',
    'scriptVersion',
    'status',
    'usageModel',
  ],
};

export function getOrderByOptions(
  name: string,
  aggregation: QueryModelMetricsAggregation,
  dimensions: string[],
): string[] {
  const options = [];

  const metricName = name.split('_')[name.split('_').length - 1];
  if (aggregation === 'count') {
    options.push(`count_ASC`);
    options.push(`count_DESC`);
  } else if (aggregation === 'avg') {
    options.push(`avg_${metricName}_ASC`);
    options.push(`avg_${metricName}_DESC`);
  } else if (aggregation === 'sum') {
    options.push(`sum_${metricName}_ASC`);
    options.push(`sum_${metricName}_DESC`);
  } else if (aggregation.startsWith('P')) {
    options.push(`quantiles_${metricName}${aggregation}_ASC`);
    options.push(`quantiles_${metricName}${aggregation}_DESC`);
  }

  for (const dimension of dimensions) {
    options.push(`${dimension}_ASC`);
    options.push(`${dimension}_DESC`);
  }

  return options;
}
