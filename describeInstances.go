// Mock EC2 Client
// Copyright (C) 2017+ Jonathan Gold and the project contributors
// Written by Jonathan Gold <info@jonathangold.ca> and the project
// contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package ec2mock

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func (c *MockEc2Client) DescribeInstances(in *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	// simulate the time it takes for aws to return
	time.Sleep(DescribeDelay * time.Second)
	var reservations []*ec2.Reservation
	for _, res := range c.Reservations {
		var reservation ec2.Reservation
		for _, inst := range res.Instances {
			instanceID := true
			stateFilter := true
			tagFilter := true
			for _, id := range in.InstanceIds {
				if aws.StringValue(id) != aws.StringValue(inst.InstanceId) {
					instanceID = false
				}
			}
			for _, filter := range in.Filters {
				if aws.StringValue(filter.Name) == InstanceStateNameFilterName {
					stateFilter = false
					for _, val := range filter.Values {
						if aws.StringValue(inst.State.Name) == aws.StringValue(val) {
							stateFilter = true
						}
					}
				}
			}
			for _, tag := range inst.Tags {
				for _, filter := range in.Filters {
					for _, val := range filter.Values {
						if aws.StringValue(filter.Name) == TagPrefix+aws.StringValue(tag.Key) {
							if aws.StringValue(val) != aws.StringValue(tag.Value) {
								tagFilter = false
							}
						}
					}
				}
			}
			if instanceID && stateFilter && tagFilter {
				reservation.Instances = append(reservation.Instances, inst)
			}
		}
		reservations = append(reservations, &reservation)
		break
	}
	return &ec2.DescribeInstancesOutput{
		Reservations: reservations,
	}, nil
}
