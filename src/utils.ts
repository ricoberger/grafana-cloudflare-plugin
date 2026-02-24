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
];

export const getAggregationOptions = (
  name: string,
): QueryModelMetricsAggregation[] | undefined => {
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

  return undefined;
};

export const filtersOptions: Record<string, string[]> = {
  httpRequests: [
    '-',
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
};

export const getDimensionsOptions = (name: string): string[] => {
  if (name.startsWith('httpRequests_overview_')) {
    return dimensionsOptions['httpRequestsOverview'];
  }

  if (name.startsWith('httpRequests_')) {
    return dimensionsOptions['httpRequests'];
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
};

export const getOrderByOptions = (
  name: string,
  aggregation: QueryModelMetricsAggregation,
  dimensions: string[],
): string[] => {
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
  }

  for (const dimension of dimensions) {
    options.push(`${dimension}_ASC`);
    options.push(`${dimension}_DESC`);
  }

  return options;
};
