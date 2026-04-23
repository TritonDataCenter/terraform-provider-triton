/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2021 Joyent, Inc.
 * Copyright 2022 MNX Cloud, Inc.
 * Copyright 2026 Edgecast Cloud LLC.
 */

package triton

import (
	"sort"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
)

type imageSort []cloudapi.Image

func sortImages(images []cloudapi.Image) []cloudapi.Image {
	sortedImages := images
	sort.Sort(sort.Reverse(imageSort(sortedImages)))
	return sortedImages
}

func (a imageSort) Len() int {
	return len(a)
}

func (a imageSort) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a imageSort) Less(i, j int) bool {
	itime := a[i].PublishedAt
	jtime := a[j].PublishedAt
	if itime == nil && jtime == nil {
		return false
	}
	if itime == nil {
		return true
	}
	if jtime == nil {
		return false
	}
	return itime.Unix() < jtime.Unix()
}
