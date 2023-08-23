import { PANEL_TYPES } from 'constants/queryBuilder';
import { useDashboard } from 'providers/Dashboard/Dashboard';
import { useCallback, useState } from 'react';
import { Layout } from 'react-grid-layout';

import GraphLayoutContainer from './GridCardLayout';

function GridGraph(): JSX.Element {
	const { selectedDashboard } = useDashboard();

	const { data } = selectedDashboard || {};
	const { widgets } = data || {};

	const [layouts, setLayout] = useState<Layout[]>(
		selectedDashboard?.data?.layout || [],
	);

	const { handleToggleDashboardSlider } = useDashboard();

	const onEmptyWidgetHandler = useCallback(() => {
		handleToggleDashboardSlider(true);

		setLayout((preLayout) => [
			{
				i: PANEL_TYPES.EMPTY_WIDGET,
				w: 6,
				x: 0,
				h: 2,
				y: 0,
			},
			...(preLayout || []),
		]);
	}, [handleToggleDashboardSlider]);

	return (
		<GraphLayoutContainer
			layouts={layouts}
			onAddPanelHandler={onEmptyWidgetHandler}
			widgets={widgets}
			setLayout={setLayout}
		/>
	);
}

export default GridGraph;
