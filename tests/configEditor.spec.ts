import { expect, test } from '@grafana/plugin-e2e';

test('smoke: should render config editor', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });
  await expect(page.getByTestId('authentication-method')).toBeVisible();
});
