import { DataSourcePlugin } from '@grafana/data';

import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { DataSource } from './datasource';
import { Options, Query } from './types';

export const plugin = new DataSourcePlugin<DataSource, Query, Options>(
  DataSource,
)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
