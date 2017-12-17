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

func (c *MockEc2Client) RunInstances(in *ec2.RunInstancesInput) (*ec2.Reservation, error) {
	// TODO: this should technically support running more than one instance
	reservation := &ec2.Reservation{}
	for i := 0; i < int(aws.Int64Value(in.MaxCount)); i++ {
		reservation.Instances = append(reservation.Instances, &ec2.Instance{
			InstanceId: aws.String(string(c.InstanceCount)),
			State: &ec2.InstanceState{
				Name: aws.String(ec2.InstanceStateNamePending),
			},
		})
	}
	// launch a goroutine to change the state after a short delay. We don't
	// have a waitgroup because we need to return right away to match the api.
	c.Reservations = append(c.Reservations, reservation)
	go func() {
		time.Sleep(time.Second * StateChangeDelay)
		for i := 0; i < len(reservation.Instances); i++ {
			reservation.Instances[i].State.Name = aws.String(ec2.InstanceStateNameRunning)
			c.InstanceCount++
		}
	}()
	return reservation, nil
}
