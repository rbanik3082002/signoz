import { Row } from 'antd';
import getLocalStorageKey from 'api/browser/localstorage/get';
import setLocalStorageKey from 'api/browser/localstorage/set';
import { LOCALSTORAGE } from 'constants/localStorage';
import { useUpdateDashboard } from 'hooks/dashboard/useUpdateDashboard';
import { useNotifications } from 'hooks/useNotifications';
import { defaultTo } from 'lodash-es';
import { useDashboard } from 'providers/Dashboard/Dashboard';
import { memo, useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { AppState } from 'store/reducers';
import { Dashboard, IDashboardVariable } from 'types/api/dashboard/getAll';
import AppReducer from 'types/reducer/app';

import { convertVariablesToDbFormat } from './util';
// import { convertVariablesToDbFormat } from './util';
import VariableItem from './VariableItem';

function DashboardVariableSelection(): JSX.Element | null {
	const {
		selectedDashboard,
		setSelectedDashboard,
		dashboardId,
	} = useDashboard();

	const { data } = selectedDashboard || {};

	const { variables } = data || {};

	const [update, setUpdate] = useState<boolean>(false);
	const [lastUpdatedVar, setLastUpdatedVar] = useState<string>('');

	const [variablesTableData, setVariablesTableData] = useState<any>([]);

	// const { role } = useSelector<AppState, AppReducer>((state) => state.app);

	useEffect(() => {
		if (variables) {
			const tableRowData = [];

			// eslint-disable-next-line no-restricted-syntax
			for (const [key, value] of Object.entries(variables)) {
				const { id } = value;

				tableRowData.push({
					key,
					name: key,
					...variables[key],
					id,
				});
			}

			tableRowData.sort((a, b) => a.order - b.order);

			setVariablesTableData(tableRowData);
		}
	}, [variables]);

	const onVarChanged = (name: string): void => {
		setLastUpdatedVar(name);
		setUpdate(!update);
	};

	const updateMutation = useUpdateDashboard();
	const { notifications } = useNotifications();

	const updateVariables = (
		name: string,
		updatedVariablesData: Dashboard['data']['variables'],
	): void => {
		if (!selectedDashboard) {
			return;
		}

		updateMutation.mutateAsync(
			{
				...selectedDashboard,
				data: {
					...selectedDashboard.data,
					variables: updatedVariablesData,
				},
			},
			{
				onSuccess: (updatedDashboard) => {
					if (updatedDashboard.payload) {
						setSelectedDashboard(updatedDashboard.payload);
					}
				},
				onError: () => {
					notifications.error({
						message: `Error updating ${name} variable`,
					});
				},
			},
		);
	};

	const onValueUpdate = (
		name: string,
		id: string,
		value: IDashboardVariable['selectedValue'],
		allSelected: boolean,
	): void => {
		if (id) {
			const newVariablesArr = variablesTableData.map(
				(variable: IDashboardVariable) => {
					const variableCopy = { ...variable };

					if (variableCopy.id === id) {
						variableCopy.selectedValue = value;
						variableCopy.allSelected = allSelected;
					}

					return variableCopy;
				},
			);

			const allDashboardVariablesFromLocalStorageString = getLocalStorageKey(
				LOCALSTORAGE.DASHBOARD_VARIABLES,
			);

			let allDashboardsFromLocalStorage = {};
			let currentDashboardVariablesFromLocalStorage = {};

			if (allDashboardVariablesFromLocalStorageString === null) {
				setLocalStorageKey(
					LOCALSTORAGE.DASHBOARD_VARIABLES,
					JSON.stringify({
						[dashboardId]: {},
					}),
				);
			} else {
				try {
					allDashboardsFromLocalStorage = JSON.parse(
						allDashboardVariablesFromLocalStorageString,
					);
					// currentDashboardVariablesFromLocalStorage = JSON.parse(
					// 	allDashboardVariablesFromLocalStorage,
					// )[dashboardId];
				} catch {
					allDashboardsFromLocalStorage = {};
				}
			}
			currentDashboardVariablesFromLocalStorage = defaultTo(
				allDashboardsFromLocalStorage?.[dashboardId],
				{},
			);
			currentDashboardVariablesFromLocalStorage[id] = {
				selectedValue: value,
				allSelected,
			};

			allDashboardsFromLocalStorage = {
				...allDashboardsFromLocalStorage,
				[dashboardId]: {
					...currentDashboardVariablesFromLocalStorage,
				},
			};

			setLocalStorageKey(
				LOCALSTORAGE.DASHBOARD_VARIABLES,
				JSON.stringify(allDashboardsFromLocalStorage),
			);

			const variables = convertVariablesToDbFormat(newVariablesArr);

			setSelectedDashboard({
				...selectedDashboard,
				data: {
					...selectedDashboard?.data,
					variables: {
						...variables,
					},
				},
			});

			// if (role !== 'VIEWER' && selectedDashboard) {
			// 	updateVariables(name, variables);
			// }
			onVarChanged(name);

			setUpdate(!update);
		}
	};

	if (!variables) {
		return null;
	}

	const orderBasedSortedVariables = variablesTableData.sort(
		(a: { order: number }, b: { order: number }) => a.order - b.order,
	);

	return (
		<Row>
			{orderBasedSortedVariables &&
				Array.isArray(orderBasedSortedVariables) &&
				orderBasedSortedVariables.length > 0 &&
				orderBasedSortedVariables.map((variable) => (
					<VariableItem
						key={`${variable.name}${variable.id}}${variable.order}`}
						existingVariables={variables}
						lastUpdatedVar={lastUpdatedVar}
						variableData={{
							name: variable.name,
							...variable,
							change: update,
						}}
						onValueUpdate={onValueUpdate}
					/>
				))}
		</Row>
	);
}

export default memo(DashboardVariableSelection);
